//
// @project geniusrabbit::xrpc 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package xrpc

import (
	"math/rand"
	"sync"
	"testing"
)

const alphabet = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/.`

var (
	mx       sync.Mutex
	keys     []string
	keys2    [][]byte
	testMap  map[string]int
	testTree pathTree
)

// go test -v -bench=. -benchmem -benchtime=10s github.com/geniusrabbit/billing/xrpc
// goos: darwin
// goarch: amd64
// pkg: github.com/geniusrabbit/billing/xrpc
// BenchmarkMap-4          100000000              132 ns/op               0 B/op          0 allocs/op
// BenchmarkTree-4         500000000               37.2 ns/op             0 B/op          0 allocs/op
// PASS
// ok      github.com/geniusrabbit/billing/xrpc     35.664s

func initData() {
	testMap = map[string]int{}
	testTree = newTree()

	for i := 0; i < 1000; i++ {
		name := randomBytes(10 + rand.Intn(100))
		keys = append(keys, string(name))
		keys2 = append(keys2, name)
		testTree.Add(name, nil)
		testMap[string(name)] = i
	}
}

func BenchmarkMap(b *testing.B) {
	initData()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			mx.Lock()
			_ = testMap[keys[i%len(keys)]]
			mx.Unlock()
		}
	})
}

func BenchmarkTree(b *testing.B) {
	initData()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			_ = testTree.Node(keys2[i%len(keys2)])
		}
	})
}

func TestTree(t *testing.T) {
	testTree := newTree()
	actions := []string{"whois", "predict", "predict_price", "device", "geo", "check"}

	for _, act := range actions {
		testTree.Add([]byte(act), nil)
	}

	for _, act := range actions {
		if node := testTree.Node([]byte(act)); node == nil {
			t.Fail()
		}
	}
}

func randomBytes(length int) (b []byte) {
	for i := 0; i < length; i++ {
		b = append(b, alphabet[rand.Intn(len(alphabet)-1)])
	}
	return
}
