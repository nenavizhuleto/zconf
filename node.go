package zconf

import (
	"context"
	"path/filepath"
)

type Node struct {
	conf *Config

	Path string
	Data []byte

	cbs callbacks
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

func (n *Node) Watch(ctx context.Context) error {
	return n.conf.watchNode(ctx, n.Path, n.cbs)
}

func (n *Node) OnChange(callback func(data []byte)) {
	n.cbs.changed = func(_ string, data []byte) { callback(data) }
}

func (n *Node) OnDelete(callback func(data []byte)) {
	n.cbs.deleted = func(_ string, data []byte) { callback(data) }
}

func (n *Node) OnChildren(callback func(children []string, new []string)) {
	n.cbs.children = func(_ string, children, new []string) { callback(children, new) }
}
