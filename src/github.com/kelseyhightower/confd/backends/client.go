package backends

import (
	"errors"

	"github.com/kelseyhightower/confd/backends/consul"
	"github.com/kelseyhightower/confd/backends/dynamodb"
	"github.com/kelseyhightower/confd/backends/env"
	"github.com/kelseyhightower/confd/backends/etcd"
	"github.com/kelseyhightower/confd/backends/redis"
	"github.com/kelseyhightower/confd/backends/stackengine"
	"github.com/kelseyhightower/confd/backends/fs"
	"github.com/kelseyhightower/confd/backends/redis"
	"github.com/kelseyhightower/confd/backends/zookeeper"
	"github.com/kelseyhightower/confd/config"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

// New is used to create a storage client based on our configuration.
func New(bc config.BackendConfig) (StoreClient, error) {
	switch bc.Type() {
	case "consul":
		return consul.NewConsulClient(bc.(*config.ConsulBackendConfig))
	case "env":
		return env.NewEnvClient(bc.(*config.EnvBackendConfig))
	case "etcd":
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		return etcd.NewEtcdClient(bc.(*config.EtcdBackendConfig))
	case "redis":
		return redis.NewRedisClient(bc.(*config.RedisBackendConfig))
	case "zookeeper":
		return zookeeper.NewZookeeperClient(bc.(*config.ZookeeperBackendConfig))
	case "dynamodb":
		return dynamodb.NewDynamoDBClient(bc.(*config.DynamoDBBackendConfig))
	case "stackengine":
		return stackengine.NewStackEngineClient(backendNodes, config.Scheme, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.AuthToken)
	case "fs":
		return fs.NewFsClient(bc.(*config.FsBackendConfig))
	}
	return nil, errors.New("Invalid backend")
}
