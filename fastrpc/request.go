//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fastrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/valyala/fastrpc/tlv"
)

type message struct {
	ID      string            `json:"id"`
	Timeout time.Duration     `json:"timeout"`
	Headers map[string][]byte `json:"headers"`
	Data    json.RawMessage   `json:"data"`
}

type request struct {
	msg    message
	ctx    context.Context
	reqCtx *tlv.RequestCtx
}

// ID of request
func (r *request) ID() []byte {
	return []byte(r.msg.ID)
}

// Action name
func (r *request) Action() []byte {
	return r.reqCtx.Request.Name()
}

// Timeout value
func (r *request) Timeout() time.Duration {
	return r.msg.Timeout
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
	return r.reqCtx
}

// Headers from request
func (r *request) Headers() map[string][]byte {
	return r.msg.Headers
}

// Bind message to object or structure
func (r *request) Bind(target interface{}) error {
	if r.msg.Data == nil {
		if err := json.Unmarshal(r.reqCtx.Request.Value(), &r.msg); err != nil {
			return err
		}
	}
	return json.Unmarshal(r.msg.Data, target)
}

// Send message as response
func (r *request) Send(msg interface{}) (err error) {
	var buff bytes.Buffer

	if err = json.NewEncoder(&buff).Encode(msg); err != nil {
		return err
	}

	_, err = r.reqCtx.Write(buff.Bytes())
	return err
}
