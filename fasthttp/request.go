//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fasthttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

type request struct {
	id      []byte
	action  []byte
	timeout time.Duration
	headers map[string][]byte
	data    []byte
	ctx     context.Context
	fastCtx *fasthttp.RequestCtx
}

// ID of request
func (r *request) ID() []byte {
	return r.id
}

// Action name
func (r *request) Action() []byte {
	return r.action
}

// Timeout value
func (r *request) Timeout() time.Duration {
	return r.timeout
}

// UpdateHeaders of request from source request context
func (r *request) UpdateHeaders() {
	r.headers = map[string][]byte{}
	r.fastCtx.Request.Header.VisitAll(func(key, value []byte) {
		r.headers[string(key)] = value
	})
}

// External Context
func (r *request) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

// SetContext new object
func (r *request) SetContext(ctx context.Context) {
	r.ctx = ctx
}

// Source of request used for processing this methods
func (r *request) Source() interface{} {
	return r.fastCtx
}

// Headers from request
func (r *request) Headers() map[string][]byte {
	return r.headers
}

// Bind message to object or structure
func (r *request) Bind(target interface{}) error {
	return json.Unmarshal(r.data, target)
}

// Send message as response
func (r *request) Send(msg interface{}) error {
	var (
		buff bytes.Buffer
		err  = json.NewEncoder(&buff).Encode(msg)
	)
	if err != nil {
		return err
	}

	r.fastCtx.Response.Reset()
	r.fastCtx.SetStatusCode(http.StatusOK)
	r.fastCtx.SetBody(buff.Bytes())

	return err
}
