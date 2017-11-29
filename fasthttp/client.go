//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fasthttp

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/demdxx/gocast"
	"github.com/geniusrabbit/xrpc"
	"github.com/valyala/fasthttp"
)

// Client implementation
type Client struct {
	hostname string
	client   *fasthttp.HostClient
}

// NewClient object connector
func NewClient(hostname string, client ...*fasthttp.HostClient) xrpc.Client {
	var c *fasthttp.HostClient
	if len(client) < 1 {
		c = &fasthttp.HostClient{
			Addr:                hostname,
			Name:                "fasthttp-client",
			Dial:                nil,
			DialDualStack:       true,
			MaxIdleConnDuration: 300 * time.Second,
			ReadBufferSize:      256 * 1024,
			WriteBufferSize:     256 * 1024,
		}
	} else {
		c = client[0]
		c.Addr = hostname
	}

	if u, _ := url.Parse(hostname); u == nil || (u.Scheme != "http" && u.Scheme != "https") {
		hostname = "http://" + hostname
	}

	return &Client{hostname: hostname, client: c}
}

// Send message to service
func (c *Client) Send(msg xrpc.Message) xrpc.Response {
	var (
		req  = fasthttp.AcquireRequest()
		resp = fasthttp.AcquireResponse()
	)

	req.ResetBody()
	req.SetRequestURI(c.hostname + "/" + msg.Action)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")

	for key, val := range msg.Headers {
		req.Header.Set(key, gocast.ToString(val))
	}

	if len(msg.ID) > 0 {
		req.Header.Set(XServiceRequestID, msg.ID)
	}

	if msg.Timeout > 0 {
		req.Header.Set(XServiceTimeout, strconv.FormatInt(int64(msg.Timeout), 10))
	}

	if err := json.NewEncoder(req.BodyWriter()).Encode(msg.Data); err != nil {
		return &Response{err: err}
	}

	if msg.Timeout <= 0 {
		return &Response{resp: resp, err: c.client.Do(req, resp)}
	}

	return &Response{
		resp: resp,
		err:  c.client.DoTimeout(req, resp, msg.Timeout),
	}
}
