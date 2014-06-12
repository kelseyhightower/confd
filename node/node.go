package node

type Node struct {
	Key   string
	Value interface{}
}

type Directory map[string][]Node

func NewDirectory() Directory {
	d := make(Directory)
	return d
}

// Add.
func (d Directory) Add(key string, node Node) {
	if d[key] == nil {
		d[key] = make([]Node, 0)
	}
	d[key] = append(d[key], node)
}

// Get.
func (d Directory) Get(key string) []Node {
	return d[key]
}

// Set.
func (d Directory) Set(key string, nodes []Node) {
	d[key] = nodes
}
