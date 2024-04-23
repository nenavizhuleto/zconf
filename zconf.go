package zconf

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/go-zookeeper/zk"
)

var (
	DefaultSessionTimeout = time.Second
)

type NodeDataCallback func(path string, data []byte)
type NodeCallback func(path string)

type Config struct {
	zcon *zk.Conn

	// callbacks
	data_cb NodeDataCallback
	node_cb NodeCallback
}

func New(servers []string) (*Config, error) {
	connection, _, err := zk.Connect(servers, DefaultSessionTimeout)
	if err != nil {
		return nil, err
	}

	return &Config{
		zcon: connection,
	}, nil
}

func (c *Config) OnNodeDataChanged(callback NodeDataCallback) {
	c.data_cb = callback
}

func (c *Config) OnNodeChanged(callback NodeCallback) {
	c.node_cb = callback
}

func (c *Config) WatchNode(nodepath string) error {
	for {
		body, _, w, err := c.zcon.GetW(nodepath)
		if err != nil {
			return err
		}

		if c.node_cb != nil {
			c.node_cb(nodepath)
		}

		if c.data_cb != nil {
			c.data_cb(nodepath, body)
		}

		event := <-w

		if event.Type == zk.EventNodeDeleted {
			return nil
		}
	}
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
