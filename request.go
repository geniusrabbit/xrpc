//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

import (
	"context"
	"time"
)

// Request describes input data
type Request interface {
	// ID of request
	ID() []byte

	// Action name
	Action() []byte

	// Timeout value
	Timeout() time.Duration

	// External Context
	Context() context.Context

	// SetContext new object
	SetContext(ctx context.Context)

	// Source of request used for processing this methods
	Source() interface{}

	// Bind message to object or structure
	Bind(target interface{}) error

	// Send message as response
	Send(msg interface{}) error
}
