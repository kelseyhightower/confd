package backends

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/log"
	confdplugin "github.com/kelseyhightower/confd/plugin"
)

// New is used to create a storage client based on our configuration.
func New(config Config) (confd.Database, error) {
	config.Backend = "env"
	// backendNodes := config.BackendNodes
	// log.Info("Backend nodes set to " + strings.Join(backendNodes, ", "))
	// switch config.Backend {
	// case "consul":
	// 	return consul.New(config.BackendNodes, config.Scheme,
	// 		config.ClientCert, config.ClientKey,
	// 		config.ClientCaKeys)
	// case "etcd":
	// 	// Create the etcd client upfront and use it for the life of the process.
	// 	// The etcdClient is an http.Client and designed to be reused.
	// 	return etcd.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	// case "zookeeper":
	// 	return zookeeper.NewZookeeperClient(backendNodes)
	// case "rancher":
	// 	return rancher.NewRancherClient(backendNodes)
	// case "redis":
	// 	return redis.NewRedisClient(backendNodes, config.ClientKey)
	// case "env":
	// 	return env.NewEnvClient()
	// case "vault":
	// 	vaultConfig := map[string]string{
	// 		"app-id":   config.AppID,
	// 		"user-id":  config.UserID,
	// 		"username": config.Username,
	// 		"password": config.Password,
	// 		"token":    config.AuthToken,
	// 		"cert":     config.ClientCert,
	// 		"key":      config.ClientKey,
	// 		"caCert":   config.ClientCaKeys,
	// 	}
	// 	return vault.New(backendNodes[0], config.AuthType, vaultConfig)
	// case "dynamodb":
	// 	table := config.Table
	// 	log.Info("DynamoDB table set to " + table)
	// 	return dynamodb.NewDynamoDBClient(table)
	// case "stackengine":
	// 	return stackengine.NewStackEngineClient(backendNodes, config.Scheme, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.AuthToken)
	// }
	// return nil, errors.New("Invalid backend")
	plugins, err := Discover()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info("Discovered: %s", plugins)
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: confdplugin.HandshakeConfig,
		Plugins:         confdplugin.PluginMap,
		Cmd:             pluginCmd(plugins[config.Backend]),
	})
	// defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Request the plugin
	log.Info("Requesting plugin")
	raw, err := rpcClient.Dispense("database")
	if err != nil {
		log.Fatal(err.Error())
	}
	return raw.(confd.Database), nil
}
