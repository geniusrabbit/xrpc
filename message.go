//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

import (
	"time"
)

// Message object
type Message struct {
	ID      string
	Action  string
	Timeout time.Duration
	Headers map[string]interface{}
	Data    interface{}
}
