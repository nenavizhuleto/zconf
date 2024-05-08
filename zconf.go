package zconf

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"time"

	"github.com/go-zookeeper/zk"
)

var (
	DefaultSessionTimeout = time.Second
)

type NodeCallback func(path string, data []byte)

type Config struct {
	zcon *zk.Conn

	binding map[string]NodeCallback
	// callbacks
	changed NodeCallback
	deleted NodeCallback
}

func New(servers []string) (*Config, error) {
	connection, _, err := zk.Connect(servers, DefaultSessionTimeout)
	if err != nil {
		return nil, err
	}

	return &Config{
		zcon:    connection,
		binding: make(map[string]NodeCallback),
	}, nil
}

func (c *Config) OnNodeChanged(callback NodeCallback) {
	c.changed = callback
}

func (c *Config) OnNodeDeleted(callback NodeCallback) {
	c.deleted = callback
}

type Node struct {
	conf *Config

	Path string
	Data []byte
}

func (n *Node) Children() (nodes []Node, err error) {
	children, _, err := n.conf.zcon.Children(n.Path)
	if err != nil {
		return
	}

	for _, child := range children {
		node, err := n.conf.Get(filepath.Join(n.Path, child))
		if err != nil {
			continue
		}

		nodes = append(nodes, node)
	}

	return
}

func (c *Config) Get(nodepath string) (node Node, err error) {
	data, _, err := c.zcon.Get(nodepath)
	if err != nil {
		return
	}

	node = Node{
		conf: c,

		Path: nodepath,
		Data: data,
	}

	return
}

func (c *Config) WatchNode(nodepath string) error {
	for {
		body, _, w, err := c.zcon.GetW(nodepath)
		if err != nil {
			return err
		}

		if c.changed != nil {
			c.changed(nodepath, body)
		}

		if callback, ok := c.binding[nodepath]; ok {
			callback(nodepath, body)
		}

		event := <-w

		if event.Type == zk.EventNodeDeleted {
			c.deleted(nodepath, body)
			return nil
		}
	}
}

func (c *Config) Bind(path string, callback NodeCallback) {
	c.binding[path] = callback
}

func (c *Config) WatchPath(path string) error {

	go c.WatchNode(path)

	children, _, w, err := c.zcon.ChildrenW(path)
	if err != nil {
		return err
	}

	for _, child := range children {
		go c.WatchPath(filepath.Join(path, child))
	}

	for {
		for event := range w {
			if event.Type == zk.EventNodeChildrenChanged {
				new_children, _, err := c.zcon.Children(path)
				if err != nil {
					continue
				}

				for _, child := range new_children {
					if slices.Contains(children, child) {
						continue
					}

					go c.WatchNode(filepath.Join(path, child))
				}

				children = new_children
			}
		}
	}
}

func (c *Config) Dump(path string, value any) error {

	exists, stat, err := c.zcon.Exists(path)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if exists {
		_, err := c.zcon.Set(path, data, stat.Version)
		if err != nil {
			return err
		}
	} else {
		_, err := c.zcon.Create(path, data, zk.FlagTTL, zk.WorldACL(zk.PermAll))
		if err != nil {
			return err
		}
	}

	return nil
}
