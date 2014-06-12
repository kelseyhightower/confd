package backends

import (
	"errors"
	"strings"

	"github.com/kelseyhightower/confd/backends/consul"
	"github.com/kelseyhightower/confd/backends/env"
	"github.com/kelseyhightower/confd/backends/etcd/etcdutil"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]interface{}, error)
}

// New is used to create a storage client based on our configuration.
func New(backend string) (StoreClient, error) {
	if backend == "" {
		backend = "etcd"
	}
	switch backend {
	case "consul":
		log.Notice("Consul address set to " + config.ConsulAddr())
		return consul.NewConsulClient(config.ConsulAddr())
	case "etcd":
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		log.Notice("etcd nodes set to " + strings.Join(config.EtcdNodes(), ", "))
		return etcdutil.NewEtcdClient(config.EtcdNodes(), config.ClientCert(), config.ClientKey(), config.ClientCaKeys())
	case "env":
		return env.NewEnvClient()
	}
	return nil, errors.New("Invalid backend")
}
