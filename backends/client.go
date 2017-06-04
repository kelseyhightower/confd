package backends

import (
	"os/exec"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/backends/commons"
	"github.com/kelseyhightower/confd/log"
)

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"env": &commons.StoreClientPlugin{},
}

// New is used to create a storage client based on our configuration.
func New(config Config) (commons.StoreClient, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("/Users/oleksa/go/src/github.com/kelseyhightower/confd/bin/plugins"),
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("env")
	if err != nil {
		log.Fatal(err.Error())
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	return raw.(commons.StoreClient), nil
}
