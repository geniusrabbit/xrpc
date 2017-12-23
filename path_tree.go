//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

import (
	"errors"
	"sort"
)

var (
	errPathIsEmpty = errors.New("Path is empty")
)

type peaceType [3]byte

type pathTree interface {
	Add(path []byte, action Action) error
	Node(path []byte) *pathTreeNode
}

func newTree() pathTree {
	return &pathTreeNode{}
}

type pathTreeNode struct {
	peace  peaceType
	Nodes  []*pathTreeNode
	Action Action
}

func (n *pathTreeNode) Add(path []byte, action Action) error {
	head, tail, err := pathPeace(path)
	if err != nil {
		return err
	}

	node := n.getOrCreateNode(head, true)
	if len(tail) < 1 {
		node.Action = action
		return nil
	}

	return node.Add(tail, action)
}

func (n *pathTreeNode) Node(path []byte) *pathTreeNode {
	head, tail, err := pathPeace(path)
	if err != nil {
		return nil
	}

	node := n.getOrCreateNode(head, false)
	if node == nil {
		return nil
	}

	if len(tail) < 1 {
		return node
	}

	return node.Node(tail)
}

func (n *pathTreeNode) getOrCreateNode(peace peaceType, create bool) (node *pathTreeNode) {
	i := sort.Search(len(n.Nodes), func(i int) bool {
		for ix, c := range n.Nodes[i].peace {
			if c > peace[ix] {
				return true
			} else if c < peace[ix] {
				break
			}
			if ix+1 == len(peace) {
				return peace[ix] == c
			}
		}
		return false
	})

	if i >= 0 && i < len(n.Nodes) && n.Nodes[i].peace == peace {
		return n.Nodes[i]
	}

	if create {
		node = &pathTreeNode{peace: peace}
		n.Nodes = append(n.Nodes, node)

		sort.Slice(n.Nodes, func(i, j int) bool {
			for ix, c := range n.Nodes[i].peace {
				if c2 := n.Nodes[j].peace[ix]; c < c2 {
					return true
				} else if c > c2 {
					break
				}
			}
			return false
		})
	}
	return
}

func pathPeace(path []byte) (head peaceType, tail []byte, err error) {
	if len(path) > len(head) {
		for i := 0; i < len(head); i++ {
			head[i] = path[i]
		}
		tail = path[len(head):]
	} else if len(path) > 0 {
		for i := 0; i < len(path); i++ {
			head[i] = path[i]
		}
		for i := len(path); i < len(head); i++ {
			head[i] = 0
		}
	} else {
		err = errPathIsEmpty
	}
	return
}
