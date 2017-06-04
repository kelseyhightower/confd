package confd

// The Database interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type Database interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}
