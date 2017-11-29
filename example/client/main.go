package main

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/geniusrabbit/xrpc"
	"github.com/geniusrabbit/xrpc/fasthttp"
	"github.com/geniusrabbit/xrpc/fastrpc"
)

var (
	flagType    = flag.String("type", "http", "Client type: http, fastrpc, fastprcmulty")
	flagConnect = flag.String("connect", "0.0.0.0:20202", "Connect address")
)

type tmsg struct {
	Name string `json:"name"`
}

func main() {
	flag.Parse()

	fmt.Println("Run", *flagType, *flagConnect)

	switch *flagType {
	case "http":
		msgLoop(fasthttp.NewClient(*flagConnect))
	case "fastrpc":
		msgLoop(fastrpc.NewClient(*flagConnect))
	case "fastprcmulty":
		msgLoop(fastrpc.NewMultipleClient(10, *flagConnect))
	}
}

func msgLoop(client xrpc.Client) {
	var (
		count      int
		errorCount int
		duration   time.Duration
		msg        = xrpc.Message{
			Action:  "hello",
			Timeout: 100 * time.Millisecond,
			Data:    &tmsg{Name: "Name " + reflect.TypeOf(client).String()},
			Headers: map[string]interface{}{},
		}
	)
	for i := 0; i < 100000; i++ {
		now := time.Now()
		msg.ID = "id_" + strconv.Itoa(i)
		if resp := client.Send(msg); resp.Error() == nil {
			var tg interface{}
			resp.Bind(&tg)
			if i%1000 == 0 {
				fmt.Println("OK", time.Now().Sub(now), tg)
			}
			count++
		} else {
			fmt.Println("err != nil", resp.Error().Error())
			errorCount++
		}
		duration += time.Now().Sub(now)
	}

	fmt.Println("===================")
	fmt.Println("===", reflect.TypeOf(client).String())
	fmt.Println("Count:       ", count)
	fmt.Println("Error Count: ", errorCount)
	fmt.Println("Time:        ", duration)
	fmt.Println("Time Per Req:", duration/time.Duration(count+errorCount))
}
