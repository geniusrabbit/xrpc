//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

// Server implements paticular transport level of service
type Server interface {
	// Listen some address which could be any connection type like:
	// tcp://hostname:port or udp://... or unix://... etc.
	Listen(address string) error
}
