//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fastrpc

import (
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/geniusrabbit/xrpc"
	"github.com/valyala/fastrpc"
	"github.com/valyala/fastrpc/tlv"
)

// CompressType is a compression type used for connections.
type CompressType byte

const (
	// CompressNone disables connection compression.
	//
	// CompressNone may be used in the following cases:
	//
	//   * If network bandwidth between client and server is unlimited.
	//   * If client and server are located on the same physical host.
	//   * If other CompressType values consume a lot of CPU resources.
	//
	CompressNone = CompressType(fastrpc.CompressNone)

	// CompressFlate uses compress/flate with default
	// compression level for connection compression.
	//
	// CompressFlate may be used in the following cases:
	//
	//     * If network bandwidth between client and server is limited.
	//     * If client and server are located on distinct physical hosts.
	//     * If both client and server have enough CPU resources
	//       for compression processing.
	//
	CompressFlate = CompressType(fastrpc.CompressFlate)

	// CompressSnappy uses snappy compression.
	//
	// CompressSnappy vs CompressFlate comparison:
	//
	//     * CompressSnappy consumes less CPU resources.
	//     * CompressSnappy consumes more network bandwidth.
	//
	CompressSnappy = CompressType(fastrpc.CompressSnappy)
)

type connAddr struct {
	// Addr is the Server address to connect to.
	Addr string

	// Dial is a custom function used for connecting to the Server.
	//
	// fasthttp.Dial is used by default.
	Dial func(addr string) (net.Conn, error)
}

func (c connAddr) GetDial(dial func(addr string) (net.Conn, error)) func(addr string) (net.Conn, error) {
	if dial != nil {
		return dial
	}
	return c.Dial
}

// MultipleClient implementation
type MultipleClient struct {
	// SniffHeader is the header written to each connection established
	// to the server.
	//
	// It is expected that the server replies with the same header.
	SniffHeader string

	// ProtocolVersion is the version of RequestWriter and ResponseReader.
	//
	// The ProtocolVersion must be changed each time RequestWriter
	// or ResponseReader changes the underlying format.
	ProtocolVersion byte

	// Addrs list of the Server address to connect to
	Addrs []connAddr

	// CompressType is the compression type used for requests.
	//
	// CompressFlate is used by default.
	CompressType CompressType

	// Dial is a custom function used for connecting to the Server.
	//
	// fasthttp.Dial is used by default.
	Dial func(addr string) (net.Conn, error)

	// TLSConfig is TLS (aka SSL) config used for establishing encrypted
	// connection to the server.
	//
	// Encrypted connections may be used for transferring sensitive
	// information over untrusted networks.
	//
	// By default connection to the server isn't encrypted.
	TLSConfig *tls.Config

	// MaxPendingRequests is the maximum number of pending requests
	// the client may issue until the server responds to them.
	//
	// DefaultMaxPendingRequests is used by default.
	MaxPendingRequests int

	// MaxBatchDelay is the maximum duration before pending requests
	// are sent to the server.
	//
	// Requests' batching may reduce network bandwidth usage and CPU usage.
	//
	// By default requests are sent immediately to the server.
	MaxBatchDelay time.Duration

	// Maximum duration for full response reading (including body).
	//
	// This also limits idle connection lifetime duration.
	//
	// By default response read timeout is unlimited.
	ReadTimeout time.Duration

	// Maximum duration for full request writing (including body).
	//
	// By default request write timeout is unlimited.
	WriteTimeout time.Duration

	// ReadBufferSize is the size for read buffer.
	//
	// DefaultReadBufferSize is used by default.
	ReadBufferSize int

	// WriteBufferSize is the size for write buffer.
	//
	// DefaultWriteBufferSize is used by default.
	WriteBufferSize int

	// Prioritizes new requests over old requests if MaxPendingRequests pending
	// requests is reached.
	PrioritizeNewRequests bool

	// clients pull of fastrpc clients
	clients []*fastrpc.Client

	// clientRoundRobinIndex it's offset counter
	clientRoundRobinIndex int32

	mx sync.Mutex
}

// NewMultipleClient connector
func NewMultipleClient(clintsCount int, addr string, addrs ...string) xrpc.Client {
	cli := &MultipleClient{
		SniffHeader:           "fastrpc",
		ProtocolVersion:       0,
		Addrs:                 nil,
		CompressType:          CompressNone,
		Dial:                  nil,
		TLSConfig:             nil,
		MaxPendingRequests:    0,
		MaxBatchDelay:         0,
		ReadTimeout:           0 * time.Millisecond,
		WriteTimeout:          0 * time.Millisecond,
		ReadBufferSize:        100 * 1024,
		WriteBufferSize:       100 * 1024,
		PrioritizeNewRequests: false,
		clients:               make([]*fastrpc.Client, clintsCount),
	}
	cli.setAddrs(addr, addrs...)
	return cli
}

// Send message to service
func (c *MultipleClient) Send(msg xrpc.Message) xrpc.Response {
	return sendMessage(c.client(), msg)
}

// SendBatch of messages
func (c *MultipleClient) SendBatch(msgs ...xrpc.Message) <-chan xrpc.Response {
	var (
		ch     = make(chan xrpc.Response, len(msgs))
		client = c.client()
	)
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(msgs))
		for _, msg := range msgs {
			go func(msg xrpc.Message) {
				ch <- sendMessage(client, msg)
				wg.Done()
			}(msg)
		}
		wg.Wait()
		close(ch)
	}()
	return ch
}

func (c *MultipleClient) setAddrs(addr string, addrs ...string) {
	var (
		naddr, dial = dialer(addr)
		cons        = []connAddr{
			{Addr: naddr, Dial: dial},
		}
	)

	for _, addr := range addrs {
		naddr, dial = dialer(addr)
		cons = append(cons, connAddr{Addr: naddr, Dial: dial})
	}

	c.Addrs = cons
}

func (c *MultipleClient) client() (cli *fastrpc.Client) {
	offset := atomic.LoadInt32(&c.clientRoundRobinIndex) % int32(len(c.clients))
	atomic.AddInt32(&c.clientRoundRobinIndex, 1)

	if cli = c.clients[offset]; cli == nil {
		c.mx.Lock()
		defer c.mx.Unlock()
		if cli = c.clients[offset]; cli == nil {
			con := c.Addrs[offset%int32(len(c.Addrs))]
			cli = &fastrpc.Client{
				SniffHeader:           c.SniffHeader,
				ProtocolVersion:       c.ProtocolVersion,
				NewResponse:           func() fastrpc.ResponseReader { return &tlv.Response{} },
				Addr:                  con.Addr,
				CompressType:          fastrpc.CompressType(c.CompressType),
				Dial:                  con.GetDial(c.Dial),
				TLSConfig:             c.TLSConfig,
				MaxPendingRequests:    c.MaxPendingRequests,
				MaxBatchDelay:         c.MaxBatchDelay,
				ReadTimeout:           c.ReadTimeout,
				WriteTimeout:          c.WriteTimeout,
				ReadBufferSize:        c.ReadBufferSize,
				WriteBufferSize:       c.WriteBufferSize,
				PrioritizeNewRequests: c.PrioritizeNewRequests,
			}
			c.clients[offset] = cli
		}
	}
	return
}
