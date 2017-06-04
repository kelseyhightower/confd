package env

// package main
//
// import (
// 	plugin "github.com/hashicorp/go-plugin"
// 	"github.com/kelseyhightower/confd/backends/commons"
// 	env "github.com/kelseyhightower/confd/backends/env/plugin"
// )
//
// // handshakeConfigs are used to just do a basic handshake between
// // a plugin and host. If the handshake fails, a user friendly error is shown.
// // This prevents users from executing bad plugins or executing a plugin
// // directory. It is a UX feature, not a security feature.
// var handshakeConfig = plugin.HandshakeConfig{
// 	ProtocolVersion:  1,
// 	MagicCookieKey:   "BASIC_PLUGIN",
// 	MagicCookieValue: "hello",
// }
//
// // pluginMap is the map of plugins we can dispense.
// var pluginMap = map[string]plugin.Plugin{
// 	"env": &commons.DatabasePlugin{Impl: new(env.Client)},
// }
//
// func main() {
// 	plugin.Serve(&plugin.ServeConfig{
// 		HandshakeConfig: handshakeConfig,
// 		Plugins:         pluginMap,
// 	})
// }
