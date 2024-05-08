package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/nenavizhuleto/zconf"
)

var (
	server = flag.String("server", "localhost:2182", "zookeeper server address")
	path   = flag.String("path", "/", "zookeeper node path")
)

func main() {
	flag.Parse()

	zc, err := zconf.New([]string{*server})
	if err != nil {
		panic(err)
	}

	zc.OnNodeChanged(func(path string, data []byte) {
		fmt.Println(path)
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-time.After(5 * time.Second)
		cancel()
	}()

	if err := zc.WatchNode(ctx, *path); err != nil {
		fmt.Println(err)
	}
}
