package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/nenavizhuleto/zconf"
)

var (
	server = flag.String("server", "localhost:2182", "zookeeper server address")
	path   = flag.String("path", "/", "zookeeper node path")
)

func main() {
	flag.Parse()

	c, err := zconf.New([]string{*server})
	if err != nil {
		flag.Usage()
		fmt.Println(err)
		return
	}

	c.OnNodeChanged(func(path string, data []byte) {
		fmt.Println("path", path, "data", string(data))
	})

	c.OnNodeError(func(path string, err error) {
		fmt.Println("path", path, "err", err)
	})

	if err := c.WatchPath(context.TODO(), *path); err != nil {
		log.Fatalln(err)
	}
}
