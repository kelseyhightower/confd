package confd

// Database interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type Database interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64) (uint64, error)
	Configure(map[string]interface{}) error
}

// DatabaseFactory is a function type that creates a new instance
// of a database.
type DatabaseFactory func() (Database, error)
