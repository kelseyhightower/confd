package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
)

// The constants below are the names of the plugins that can be dispensed
// from the plugin server.
const (
	DatabasePluginName = "database"
)

// HandshakeConfig is used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	// The ProtocolVersion is the version that must match between confd core
	// and confd plugins. This should be bumped whenever a change happens in
	// one or the other that makes it so that they can't safely communicate.
	// This could be e.g. adding a new interface value
	ProtocolVersion: 1,

	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "CONFD_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "xtanpmnh5nqffr256vnevap59w86p3wmxkfkcp9hx4fzf5frjc3tf3tkcczcatyd",
}

type DatabaseFunc func() confd.Database

// ServeOpts are the configurations to serve a plugin.
type ServeOpts struct {
	DatabaseFunc DatabaseFunc
}

// Serve serves a plugin. This function never returns and should be the final
// function called in the main function of the plugin.
func Serve(opts *ServeOpts) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap(opts),
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

// pluginMap returns the map[string]plugin.Plugin to use for configuring a plugin
// server or client.
func pluginMap(opts *ServeOpts) map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		DatabasePluginName: &DatabasePlugin{Impl: opts.DatabaseFunc()},
	}
}
