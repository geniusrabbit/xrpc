package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/geniusrabbit/xrpc"
	"github.com/geniusrabbit/xrpc/fasthttp"
	"github.com/geniusrabbit/xrpc/fastrpc"
)

var (
	flagType    = flag.String("type", "http", "Client type: http, httpmulty, fastrpc")
	flagConnect = flag.String("connect", "0.0.0.0:20202", "Connect address")
)

type tmsg struct {
	Name string `json:"name"`
}

func main() {
	flag.Parse()

	fmt.Println("Run example service", *flagType, *flagConnect)

	srv := xrpc.New()
	srv.Register("hello", helloHandler)

	switch *flagType {
	case "http":
		fatalError(fasthttp.NewServer(srv).Listen(*flagConnect))
	case "fastrpc":
		fatalError(fastrpc.NewServer(srv).Listen(*flagConnect))
	}
}

func helloHandler(req xrpc.Request) error {
	var msg tmsg
	if err := req.Bind(&msg); err != nil {
		return err
	}

	return req.Send(map[string]string{
		"id":  string(req.ID()),
		"msg": "Hello " + msg.Name + "!",
	})
}

func fatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
