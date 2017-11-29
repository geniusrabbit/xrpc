//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fasthttp

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/geniusrabbit/xrpc"
	"github.com/valyala/fasthttp"
)

type server struct {
	service xrpc.Service
	fastsrv fasthttp.Server
}

// NewServer default configurated server
func NewServer(service xrpc.Service) xrpc.Server {
	return &server{
		service: service,
		fastsrv: fasthttp.Server{
			Name:        "fasthttp",
			Concurrency: 1000,
		},
	}
}

// Listen some address which could be any connection type like:
// tcp://hostname:port or udp://... or unix://... etc.
func (s *server) Listen(address string) error {
	s.fastsrv.Handler = s.handler
	if strings.HasPrefix(address, "unix://") {
		return s.fastsrv.ListenAndServeUNIX(strings.TrimPrefix(address, "unix://"), 0664)
	}
	return s.fastsrv.ListenAndServe(address)
}

func (s *server) handler(ctx *fasthttp.RequestCtx) {
	var (
		tmHeader   = string(ctx.Request.Header.PeekBytes([]byte(XServiceTimeout)))
		timeout, _ = strconv.ParseInt(tmHeader, 10, 64)
		req        = &request{
			id:      ctx.Request.Header.PeekBytes([]byte(XServiceRequestID)),
			action:  bytes.TrimLeft(ctx.Path(), "/"),
			data:    ctx.Request.Body(),
			timeout: time.Duration(timeout),
			ctx:     s.requestCtx(ctx),
			fastCtx: ctx,
		}
	)

	req.UpdateHeaders()

	if err := s.service.Handle(req); err != nil {
		if err == xrpc.ErrActionNotFound {
			s.handlerNotFound(ctx)
		} else {
			s.handlerError(ctx, err)
		}
	}
}

func (s *server) handlerError(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.Reset()
	ctx.SetStatusCode(http.StatusInternalServerError)
	ctx.SetBody([]byte(`{"error":"` + strings.Replace(err.Error(), `"`, `\"`, -1) + `"}`))
}

func (s *server) handlerNotFound(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.SetStatusCode(http.StatusNotFound)
	ctx.SetBody([]byte(`{"error":"action not found"}`))
}

func (s *server) requestCtx(ctx *fasthttp.RequestCtx) context.Context {
	return context.Background()
}
