//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fastrpc

import (
	"sync"

	"github.com/geniusrabbit/xrpc"
)

var (
	requestPool = sync.Pool{New: func() interface{} {
		return new(request)
	}}
)

// BorrowReuest from pool
func BorrowReuest() xrpc.Request {
	return requestPool.Get().(xrpc.Request)
}

// ReturnRequest to pool
func ReturnRequest(req xrpc.Request) {
	if r, _ := req.(*request); r != nil {
		*r = request{}
		requestPool.Put(r)
	}
}
