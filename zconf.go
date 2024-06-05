package zconf

import (
	"context"
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

type callbacks struct {
	changed NodeCallback
	deleted NodeCallback
}

type Config struct {
	zcon *zk.Conn

	cbs callbacks
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

func (c *Config) OnNodeChanged(callback NodeCallback) {
	c.cbs.changed = callback
}

func (c *Config) OnNodeDeleted(callback NodeCallback) {
	c.cbs.deleted = callback
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

func (c *Config) watchNode(ctx context.Context, path string, cbs callbacks) error {
	for {
		body, _, w, err := c.zcon.GetW(path)
		if err != nil {
			return err
		}

		if cbs.changed != nil {
			cbs.changed(path, body)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-w:
			if event.Type == zk.EventNodeDeleted {
				if cbs.deleted != nil {
					cbs.deleted(path, body)
				}
				return nil
			}
		}
	}
}

func (c *Config) WatchNode(ctx context.Context, path string) error {
	return c.watchNode(ctx, path, c.cbs)
}

func (c *Config) WatchPath(ctx context.Context, path string) error {

	go c.WatchNode(ctx, path)

	children, _, w, err := c.zcon.ChildrenW(path)
	if err != nil {
		return err
	}

	for _, child := range children {
		go c.WatchPath(ctx, filepath.Join(path, child))
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-w:
			if event.Type == zk.EventNodeChildrenChanged {
				new_children, _, err := c.zcon.Children(path)
				if err != nil {
					continue
				}

				for _, child := range new_children {
					if slices.Contains(children, child) {
						continue
					}

					go c.WatchNode(ctx, filepath.Join(path, child))
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
