//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

import "errors"

// Service errors
var (
	ErrActionNotFound  = errors.New("Action not found")
	ErrInvalidResponse = errors.New("Invalid response")
)

// Middleware of service
type Middleware interface {
	Handle(req Request) error
}

// Service accessor interface
type Service interface {
	// Use middleware in the loop
	Use(m Middleware)

	// Register service action function
	Register(name string, fnk Action) error

	// Handle paticular request
	Handle(req Request) error
}

type service struct {
	actions      pathTree
	middlewaries []Middleware
}

// New sevice default connector
func New() Service {
	return &service{
		actions: newTree(),
	}
}

// Use middleware in the loop
func (s *service) Use(m Middleware) {
	if m != nil {
		s.middlewaries = append(s.middlewaries, m)
	}
}

// Register service action function
func (s *service) Register(name string, fnk Action) error {
	return s.actions.Add([]byte(name), fnk)
}

// Handle action
func (s *service) Handle(req Request) error {
	if node := s.actions.Node(req.Action()); node != nil {
		for _, m := range s.middlewaries {
			if err := m.Handle(req); err != nil {
				return err
			}
		}
		return node.Action(req)
	}
	return ErrActionNotFound
}
