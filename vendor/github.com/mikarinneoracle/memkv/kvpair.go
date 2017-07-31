package memkv

type KVPair struct {
	Key   string
	Value string
}

type KVPairs []KVPair

func (ks KVPairs) Len() int {
	return len(ks)
}

func (ks KVPairs) Less(i, j int) bool {
	return ks[i].Key < ks[j].Key
}

func (ks KVPairs) Swap(i, j int) {
	ks[i], ks[j] = ks[j], ks[i]
}
