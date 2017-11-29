//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

// Client interface describer
type Client interface {
	// Send message to service
	Send(msg Message) Response
}
