package backends

import (
	"strings"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/log"
	confdplugin "github.com/kelseyhightower/confd/plugin"
)

// New is used to create a storage client based on our configuration.
func New(config Config) (confd.Database, error) {
	config.Backend = "env"
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
	database := raw.(confd.Database)

	// Configure each type of database
	var c map[string]interface{}
	log.Info("Backend nodes set to " + strings.Join(config.BackendNodes, ", "))
	switch config.Backend {
	case "consul":
		c["nodes"] = config.BackendNodes
		c["scheme"] = config.Scheme
		c["key"] = config.ClientKey
		c["cert"] = config.ClientCert
		c["caCert"] = config.ClientCaKeys
	case "env":
	case "etcd":
		c["machines"] = config.BackendNodes
		c["key"] = config.ClientKey
		c["cert"] = config.ClientCert
		c["caCert"] = config.ClientCaKeys
		c["basicAuth"] = config.BasicAuth
		c["username"] = config.Username
		c["password"] = config.Password
	case "dynamodb":
		c["table"] = config.Table
		log.Info("DynamoDB table set to %s", config.Table)
	case "rancher":
		c["backendNodes"] = config.BackendNodes
	default:
		panic("Invalid backend")
	}
	database.Configure(c)
	// case "zookeeper":
	// 	return zookeeper.NewZookeeperClient(backendNodes)
	// case "redis":
	// 	return redis.NewRedisClient(backendNodes, config.ClientKey)
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
	// case "stackengine":
	// 	return stackengine.NewStackEngineClient(backendNodes, config.Scheme, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.AuthToken)
	// }

	return database, nil
}
