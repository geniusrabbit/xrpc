//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fastrpc

import (
	"encoding/json"
	"net"
	"net/url"
	"time"

	"github.com/geniusrabbit/xrpc"
	"github.com/valyala/fastrpc"
	"github.com/valyala/fastrpc/tlv"
)

// Client implementation
type Client struct {
	client *fastrpc.Client
}

// NewClient connector
func NewClient(addr string) xrpc.Client {
	addr, dial := dialer(addr)
	return &Client{
		client: &fastrpc.Client{
			SniffHeader:           "fastrpc",
			ProtocolVersion:       0,
			NewResponse:           func() fastrpc.ResponseReader { return &tlv.Response{} },
			Addr:                  addr,
			CompressType:          fastrpc.CompressNone,
			Dial:                  dial,
			TLSConfig:             nil,
			MaxPendingRequests:    0,
			MaxBatchDelay:         0,
			ReadTimeout:           0 * time.Millisecond,
			WriteTimeout:          0 * time.Millisecond,
			ReadBufferSize:        100 * 1024,
			WriteBufferSize:       100 * 1024,
			PrioritizeNewRequests: false,
		},
	}
}

// Send message to service
func (c *Client) Send(msg xrpc.Message) xrpc.Response {
	return sendMessage(c.client, msg)
}

func sendMessage(client *fastrpc.Client, msg xrpc.Message) xrpc.Response {
	var (
		req     = tlv.AcquireRequest()
		timeout = msg.Timeout
	)

	defer tlv.ReleaseRequest(req)

	if err := json.NewEncoder(req).Encode(struct {
		ID      string        `json:"id,omitempty"`
		Timeout time.Duration `json:"timeout,omitempty"`
		Headers interface{}   `json:"headers,omitempty"`
		Data    interface{}   `json:"data"`
	}{
		ID:      msg.ID,
		Timeout: msg.Timeout,
		Headers: mapOrNil(msg.Headers),
		Data:    msg.Data,
	}); err != nil {
		return &Response{err: err}
	}

	req.SetName(msg.Action)
	if msg.Timeout <= 0 {
		timeout = client.MaxBatchDelay
	}
	if timeout <= 0 {
		timeout = 100 * time.Millisecond
	}

	resp := tlv.AcquireResponse()
	return &Response{
		resp: resp,
		err:  client.DoDeadline(req, resp, time.Now().Add(timeout)),
	}
}

func mapOrNil(m map[string]interface{}) map[string]interface{} {
	if m == nil || len(m) < 1 {
		return nil
	}
	return m
}

func dialer(addr string) (string, func(addr string) (net.Conn, error)) {
	var (
		url, err = url.Parse(addr)
		scheme   string
	)
	if err != nil {
		return addr, nil
	}

	switch url.Scheme {
	case "", "tcp", "tcp4", "tcp6":
		return url.Host, nil
	default:
		scheme = url.Scheme
	}

	return url.Host, func(addr string) (net.Conn, error) {
		return net.Dial(scheme, addr)
	}
}
