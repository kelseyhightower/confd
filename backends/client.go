package backends

import (
	"fmt"
	"strconv"
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
	log.Debug("Discovered: %s", plugins)
	if _, ok := plugins[config.Backend]; ok == false {
		return nil, nil, fmt.Errorf("Plugin %s not found", config.Backend)
	}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  confdplugin.HandshakeConfig,
		Plugins:          confdplugin.PluginMap,
		Cmd:              pluginCmd(plugins[config.Backend]),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           *log.GetLogger(),
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Request the plugin
	log.Info("Requesting plugin " + confdplugin.DatabasePluginName)
	raw, err := rpcClient.Dispense(confdplugin.DatabasePluginName)
	if err != nil {
		log.Fatal(err.Error())
	}
	database := raw.(confd.Database)

	// Configure each type of database
	c := make(map[string]string)
	if config.Backend == "file" {
		log.Info("Backend source(s) set to " + config.YAMLFile)
	} else {
		log.Info("Backend source(s) set to " + strings.Join(config.BackendNodes, ", "))
	}
	switch config.Backend {
	case "consul":
		c["nodes"] = strings.Join(config.BackendNodes, ",")
		c["scheme"] = config.Scheme
		c["key"] = config.ClientKey
		c["cert"] = config.ClientCert
		c["caCert"] = config.ClientCaKeys
		c["basicAuth"] = strconv.FormatBool(config.BasicAuth)
		c["username"] = config.Username
		c["password"] = config.Password
	case "etcd", "etcdv3":
		c["machines"] = strings.Join(config.BackendNodes, ",")
		c["key"] = config.ClientKey
		c["cert"] = config.ClientCert
		c["caCert"] = config.ClientCaKeys
		c["basicAuth"] = strconv.FormatBool(config.BasicAuth)
		c["username"] = config.Username
		c["password"] = config.Password
	case "dynamodb":
		c["table"] = config.Table
		log.Info("DynamoDB table set to %s", config.Table)
	case "rancher":
		c["backendNodes"] = strings.Join(config.BackendNodes, ",")
	case "zookeeper":
		c["machines"] = strings.Join(config.BackendNodes, ",")
	case "redis":
		c["machines"] = strings.Join(config.BackendNodes, ",")
		c["password"] = config.ClientKey
		c["separator"] = config.Separator
	case "vault":
		c["authType"] = config.AuthType
		c["address"] = config.BackendNodes[0]
		c["app-id"] = config.AppID
		c["user-id"] = config.UserID
		c["role-id"] = config.RoleID
		c["secret-id"] = config.SecretID
		c["username"] = config.Username
		c["password"] = config.Password
		c["token"] = config.AuthToken
		c["cert"] = config.ClientCert
		c["key"] = config.ClientKey
		c["caCert"] = config.ClientCaKeys
	case "file":
		c["yamlFile"] = config.YAMLFile
	}
	database.Configure(c)

	return database, client, nil
}
