//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package fastrpc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/geniusrabbit/xrpc"
	"github.com/valyala/fastrpc"
	"github.com/valyala/fastrpc/tlv"
	"github.com/valyala/tcplisten"
)

type server struct {
	service xrpc.Service
	rpc     fastrpc.Server
}

// NewServer default configurated server
func NewServer(service xrpc.Service) xrpc.Server {
	return &server{
		service: service,
		rpc: fastrpc.Server{
			SniffHeader:      "fastrpc",
			ProtocolVersion:  0,
			NewHandlerCtx:    newHandlerCtx,
			Handler:          nil,
			CompressType:     fastrpc.CompressNone,
			Concurrency:      100,
			MaxBatchDelay:    0,
			ReadTimeout:      0,
			WriteTimeout:     0,
			ReadBufferSize:   10 * 1024,
			WriteBufferSize:  10 * 1024,
			PipelineRequests: false,
		},
	}
}

// Listen some address which could be any connection type like:
// tcp://hostname:port or udp://... or unix://... etc.
func (s *server) Listen(address string) error {
	var (
		u, err   = url.Parse(address)
		listener net.Listener
	)

	if err != nil {
		return err
	}

	switch u.Scheme {
	case "tcp", "tcp4", "tcp6":
		var cfg = tcplisten.Config{ReusePort: false}
		if u.Scheme == "tcp" {
			u.Scheme = "tcp4"
		}
		if listener, err = cfg.NewListener(u.Scheme, u.Host); err != nil {
			return err
		}
		defer listener.Close()
	case "unix", "unixpacket":
		if listener, err = net.Listen(u.Scheme, u.Host); err != nil {
			return err
		}
		defer listener.Close()
	default:
		return fmt.Errorf("connection type [%s] not supported", u.Scheme)
	}

	s.rpc.Handler = s.handler
	return s.rpc.Serve(listener)
}

func (s *server) handler(tctx fastrpc.HandlerCtx) fastrpc.HandlerCtx {
	var (
		ctx = tctx.(*tlv.RequestCtx)
		req = request{
			ctx:    s.requestCtx(ctx),
			reqCtx: ctx,
		}
	)

	if err := s.service.Handle(&req); err != nil {
		if err == xrpc.ErrActionNotFound {
			s.handlerNotFound(ctx)
		} else {
			s.handlerError(ctx, err)
		}
	}
	return ctx
}

func (s *server) requestCtx(ctx fastrpc.HandlerCtx) context.Context {
	return context.Background()
}

func (s *server) handlerNotFound(ctx *tlv.RequestCtx) {
	ctx.Response.SwapValue([]byte(`{"error":"not found"}`))
}

func (s *server) handlerError(ctx *tlv.RequestCtx, err error) {
	ctx.Response.SwapValue([]byte(`{"error":"` + strings.Replace("\"", "\\\"", err.Error(), -1) + `"}`))
}

func newHandlerCtx() fastrpc.HandlerCtx {
	return &tlv.RequestCtx{
		ConcurrencyLimitErrorHandler: concurrencyLimitErrorHandler,
	}
}

func concurrencyLimitErrorHandler(ctx *tlv.RequestCtx, concurrency int) {
	ctx.Response.SwapValue([]byte(`{"error":"too many requests"}`))
}
