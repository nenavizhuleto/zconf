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

func (c *Config) WatchNode(ctx context.Context, path string) error {
	return c.watchNode(ctx, path, c.cbs)
}

func (c *Config) WatchPath(ctx context.Context, path string) error {
	node, err := c.Get(path)
	if err != nil {
		return err
	}

	node.OnChildren(func(_, new []string) {
		for _, child := range new {
			go c.WatchPath(ctx, filepath.Join(path, child))
		}
	})

	children, err := node.Children()
	if err != nil {
		return err
	}

	for _, child := range children {
		go c.WatchPath(ctx, child.Path)
	}

	return node.Watch(ctx)
}

func (c *Config) Put(path string, value any) error {
	return c.Dump(path, value)
}

// Deprecated: use Put instead
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

func (c *Config) OnNodeChanged(callback NodeCallback) {
	c.cbs.changed = callback
}

func (c *Config) OnNodeDeleted(callback NodeCallback) {
	c.cbs.deleted = callback
}

func (c *Config) OnNodeError(callback NodeErrorCallback) {
	c.cbs.error = callback
}

func (c *Config) Get(path string) (node Node, err error) {
	data, _, err := c.zcon.Get(path)
	if err != nil {
		return
	}

	node = Node{
		conf: c,

		Path: path,
		Data: data,
		cbs:  c.cbs,
	}

	return
}

func (c *Config) nodeChildrenListener(ctx context.Context, path string, cbs callbacks) error {
	for {
		children, _, cw, err := c.zcon.ChildrenW(path)
		if err != nil {
			cbs.Error(path, err)
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-cw:
			switch event.Type {
			case zk.EventNodeChildrenChanged:
				newChildren, _, err := c.zcon.Children(path)
				if err != nil {
					cbs.Error(path, err)
					return err
				}

				var onlyNewChildren []string
				for _, child := range newChildren {
					if slices.Contains(children, child) {
						continue
					}

					onlyNewChildren = append(onlyNewChildren, child)
				}

				children = newChildren

				cbs.Children(path, newChildren, onlyNewChildren)
			}
		}
	}
}

func (c *Config) nodeBodyListener(ctx context.Context, path string, cbs callbacks) error {
	for {
		body, _, w, err := c.zcon.GetW(path)
		if err != nil {
			cbs.Error(path, err)
			return err
		}

		cbs.Changed(path, body)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-w:
			switch event.Type {
			case zk.EventNodeDeleted:
				cbs.Deleted(path, body)
			}
		}
	}
}

func (c *Config) watchNode(ctx context.Context, path string, cbs callbacks) error {
	go c.nodeChildrenListener(ctx, path, cbs)
	go c.nodeBodyListener(ctx, path, cbs)
	<-ctx.Done()
	return ctx.Err()
}
