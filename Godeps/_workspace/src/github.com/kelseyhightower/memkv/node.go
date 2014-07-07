package memkv

type Node struct {
	Key   string
	Value string
}

type Nodes []Node

func (ns Nodes) Len() int {
	return len(ns)
}

func (ns Nodes) Less(i, j int) bool {
	return ns[i].Key < ns[j].Key
}

func (ns Nodes) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}
