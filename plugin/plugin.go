package plugin

import plugin "github.com/hashicorp/go-plugin"

// PluginMap should be used by clients for the map of plugins.
var PluginMap = map[string]plugin.Plugin{
	DatabasePluginName: &DatabasePlugin{},
}
