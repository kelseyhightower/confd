package backends

import (
	"errors"
	"net/url"
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
	var err error
	backendNodes := config.BackendNodes
	if config.Backend == "etcd" {
		backendNodes, err = addScheme(config.Scheme, config.BackendNodes)
		if err != nil {
			return nil, err
		}
	}
	log.Notice("Backend nodes set to " + strings.Join(backendNodes, ", "))
	switch config.Backend {
	case "consul":
		return consul.NewConsulClient(backendNodes)
	case "etcd":
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		return etcd.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys)
	case "env":
		return env.NewEnvClient()
	}
	return nil, errors.New("Invalid backend")
}

func addScheme(scheme string, nodes []string) ([]string, error) {
	ns := make([]string, 0)
	if scheme == "" {
		scheme = "http"
	}
	for _, node := range nodes {
		u, err := url.Parse(node)
		if err != nil {
			return nil, err
		}
		if u.Scheme == "" {
			u.Scheme = scheme
		}
		ns = append(ns, u.String())
	}
	return ns, nil
}
