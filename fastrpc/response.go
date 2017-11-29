//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fastrpc

import (
	"encoding/json"
	"errors"

	"github.com/geniusrabbit/xrpc"
	"github.com/valyala/fastrpc/tlv"
)

// Response wrapper
type Response struct {
	parsedError bool
	resp        *tlv.Response
	err         error
}

// Source of request used for processing this methods
func (r Response) Source() interface{} {
	return r.resp
}

// Bind message to object or structure
func (r Response) Bind(target interface{}) error {
	if r.err != nil {
		return r.err
	}
	if r.resp == nil {
		return xrpc.ErrInvalidResponse
	}
	return json.Unmarshal(r.resp.Value(), target)
}

// Error response
func (r *Response) Error() error {
	if r.err == nil && r.resp != nil {
		var err struct {
			Error string `json:"error"`
		}
		if e := json.Unmarshal(r.resp.Value(), &err); e == nil {
			if err.Error == "action not found" {
				r.err = xrpc.ErrActionNotFound
			} else if err.Error != "" {
				r.err = errors.New(err.Error)
			}
		}
		r.parsedError = true
	}
	return r.err
}
