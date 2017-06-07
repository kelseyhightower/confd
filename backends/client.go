package backends

import (
	"strings"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/log"
	confdplugin "github.com/kelseyhightower/confd/plugin"
)

// New is used to create a storage client based on our configuration.
func New(config Config) (confd.Database, *plugin.Client, error) {
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
	raw, err := rpcClient.Dispense(confdplugin.DatabasePluginName)
	if err != nil {
		log.Fatal(err.Error())
	}
	database := raw.(confd.Database)

	// Configure each type of database
	c := make(map[string]interface{})
	log.Info("Backend nodes set to " + strings.Join(config.BackendNodes, ", "))
	switch config.Backend {
	case "consul":
		c["nodes"] = config.BackendNodes
		c["scheme"] = config.Scheme
		c["key"] = config.ClientKey
		c["cert"] = config.ClientCert
		c["caCert"] = config.ClientCaKeys
	case "env":
		break
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
	case "zookeeper":
		c["machines"] = config.BackendNodes
	case "redis":
		c["machines"] = config.BackendNodes
		c["password"] = config.ClientKey
	case "vault":
		c["authType"] = config.AuthType
		c["address"] = config.BackendNodes[0]
		c["app-id"] = config.AppID
		c["user-id"] = config.UserID
		c["username"] = config.Username
		c["password"] = config.Password
		c["token"] = config.AuthToken
		c["cert"] = config.ClientCert
		c["key"] = config.ClientKey
		c["caCert"] = config.ClientCaKeys
	case "stackengine":
		c["nodes"] = config.BackendNodes
		c["cert"] = config.ClientCert
		c["key"] = config.ClientKey
		c["caCert"] = config.ClientCaKeys
		c["scheme"] = config.Scheme
		c["authToken"] = config.AuthToken
	default:
		panic("Invalid backend")
	}
	database.Configure(c)

	return database, client, nil
}
