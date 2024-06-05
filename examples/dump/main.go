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

type MyConfig struct {
	Port int    `json:"port"`
	Test string `json:"test"`
}

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

	if err := c.Put("/test", MyConfig{
		Port: 4123,
		Test: "zconf testing 2",
	}); err != nil {
		log.Fatalln(err)

	}

	c.WatchPath(context.TODO(), *path)
}
