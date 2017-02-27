package backends

import (
	"errors"
	"strings"

	"github.com/frostyslav/confd/backends/azuretablestorage"
	"github.com/frostyslav/confd/backends/consul"
	"github.com/frostyslav/confd/backends/dynamodb"
	"github.com/frostyslav/confd/backends/env"
	"github.com/frostyslav/confd/backends/etcd"
	"github.com/frostyslav/confd/backends/etcdv3"
	"github.com/frostyslav/confd/backends/fallback"
	"github.com/frostyslav/confd/backends/file"
	"github.com/frostyslav/confd/backends/rancher"
	"github.com/frostyslav/confd/backends/redis"
	"github.com/frostyslav/confd/backends/stackengine"
	"github.com/frostyslav/confd/backends/vault"
	"github.com/frostyslav/confd/backends/zookeeper"
	"github.com/frostyslav/confd/log"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string, token string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

// New is used to create a storage client based on our configuration.
func New(config Config) (StoreClient, error) {
	if config.Backend == "" {
		config.Backend = "etcd"
	}
	backendNodes := config.BackendNodes
	if config.Backend == "file" {
		log.Info("Backend source(s) set to " + config.YAMLFile)
	} else {
		log.Info("Backend source(s) set to " + strings.Join(backendNodes, ", "))
	}

	log.Info("Backend nodes set to " + strings.Join(backendNodes, ", "))
	if config.BackendFallback != "" {
		mainConfig := config
		fallbackConfig := config
		mainConfig.BackendFallback = ""
		fallbackConfig.Backend = config.BackendFallback
		fallbackConfig.BackendFallback = ""
		backendMain, err := New(mainConfig)
		if err != nil {
			return nil, err
		}
		backendFallback, err := New(fallbackConfig)
		if err != nil {
			return nil, err
		}
		return fallback.NewFallbackClient(backendMain, backendFallback)
	}

	switch config.Backend {
	case "azuretablestorage":
		return azuretablestorage.NewAzureTableStorageClient(config.Table, config.StorageAccount, config.ClientKey)
	case "consul":
		return consul.New(config.BackendNodes, config.Scheme,
			config.ClientCert, config.ClientKey,
			config.ClientCaKeys)
	case "etcd":
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		return etcd.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case "etcdv3":
		return etcdv3.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case "zookeeper":
		return zookeeper.NewZookeeperClient(backendNodes)
	case "rancher":
		return rancher.NewRancherClient(backendNodes)
	case "redis":
		return redis.NewRedisClient(backendNodes, config.ClientKey)
	case "env":
		return env.NewEnvClient()
	case "file":
		return file.NewFileClient(config.YAMLFile)
	case "vault":
		vaultConfig := map[string]string{
			"app-id":   config.AppID,
			"user-id":  config.UserID,
			"username": config.Username,
			"password": config.Password,
			"token":    config.AuthToken,
			"cert":     config.ClientCert,
			"key":      config.ClientKey,
			"caCert":   config.ClientCaKeys,
		}
		return vault.New(backendNodes[0], config.AuthType, vaultConfig)
	case "dynamodb":
		log.Info("DynamoDB table set to " + config.Table)
		log.Info("DynamoDB endpoint set to " + config.Endpoint)
		return dynamodb.NewDynamoDBClient(config.Table, config.Endpoint)
	case "stackengine":
		return stackengine.NewStackEngineClient(backendNodes, config.Scheme, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.AuthToken)
	}
	return nil, errors.New("Invalid backend")
}
