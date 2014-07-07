package backends

import (
	"errors"
	"strings"

	"github.com/kelseyhightower/confd/backends/consul"
	"github.com/kelseyhightower/confd/backends/env"
	"github.com/kelseyhightower/confd/backends/etcd"
	"github.com/kelseyhightower/confd/log"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
}

// New is used to create a storage client based on our configuration.
func New(config Config) (StoreClient, error) {
	if config.Backend == "" {
		config.Backend = "etcd"
	}
	log.Notice("Backend nodes set to " + strings.Join(config.BackendNodes, ", "))
	switch config.Backend {
	case "consul":
		return consul.NewConsulClient(config.BackendNodes)
	case "etcd":
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		return etcd.NewEtcdClient(config.BackendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys)
	case "env":
		return env.NewEnvClient()
	}
	return nil, errors.New("Invalid backend")
}
