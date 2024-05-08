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

	zc, err := zconf.New([]string{*server})
	if err != nil {
		panic(err)
	}

	node, err := zc.Get(*path)
	if err != nil {
		panic(err)
	}
	fmt.Println(node.Path)

	children, err := node.Children()
	for _, child := range children {
		fmt.Println("child: ", child.Path)
	}
}
