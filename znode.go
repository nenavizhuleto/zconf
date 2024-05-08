package zconf

import "path/filepath"

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
