package zconf

type NodeCallback func(path string, data []byte)
type NodeErrorCallback func(path string, err error)
type NodeChildrenCallback func(path string, children []string, new []string)

type callbacks struct {
	changed  NodeCallback
	deleted  NodeCallback
	children NodeChildrenCallback
	error    NodeErrorCallback
}

func (cbs callbacks) Changed(path string, data []byte) {
	if cbs.changed != nil {
		cbs.changed(path, data)
	}
}

func (cbs callbacks) Deleted(path string, data []byte) {
	if cbs.deleted != nil {
		cbs.deleted(path, data)
	}
}

func (cbs callbacks) Children(path string, children []string, new []string) {
	if cbs.children != nil {
		cbs.children(path, children, new)
	}
}

func (cbs callbacks) Error(path string, err error) {
	if cbs.error != nil {
		cbs.error(path, err)
	}
}
