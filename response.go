//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

// Response describes output data
type Response interface {
	// Source of request used for processing this methods
	Source() interface{}

	// Bind message to object or structure
	Bind(target interface{}) error

	// Error response
	Error() error
}
