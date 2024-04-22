package main

import (
	"flag"
	"fmt"

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

	c.OnNodeChanged(func(path string) {
		fmt.Println("path", path, "changed")
	})

	c.OnNodeDataChanged(func(path string, data []byte) {
		fmt.Println("path", path, "data", string(data))
	})

	c.WatchPath(*path)
}
