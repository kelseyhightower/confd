# Plugin system

The plugin system is implemented using https://github.com/hashicorp/go-plugin, which is successfully used in Terraform, Packer, Nomad and Vault. It's based on network RPC over the local network. Plugins are Go interface implementations, which maps quite good to the current backends implementation.

The core set of plugins is defined in builtin/ directory. They are compiled in the main binary.

Builtin plugins are called using special cli command:
`confd internal-plugin <pluginType> <pluginName>`

Third-party plugins should be placed either where confd is installed or where confd is invoked. In the current implementation, a third-party plugin should manage it's own config. For a discussion on how a plugin should be configured, please, look [Plugin Configuration](#plugin-configuration)

## Interface changes

The new interface which plugins implement has the new method `Configure`, which is used to decouple plugin configuration from a core of confd. This method is called to initialize a plugin with a config. A config is passed to a plugin in a form of `map[string]interface{}`. Config deserialization is done using `github.com/mitchellh/mapstructure`, which is a simple and a safe way to convert a map to a Go struct. A plugin may extract values from a map directly, but this behavior is unsafe and discouraged.

```go
type Database interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
	Configure(map[string]interface{}) error
}
```

## Plugin configuration

Each plugin should have its own config.
confd should only know which plugin to dispatch, but shouldn't know about any configuration options.

How should plugins be configured?
1. Plugins should be configured per template. Less reusable. Easy to understand. Hard to configure complex backends, which tend to have a lot of configuration options.
2. Plugins should be configured globally via a file, e.g. per plugin ~/.confd/etcd.toml or all-in-one ~/.confd.toml. Less repeatable then 1,
3. Plugins should be configured via environment variables. How to figure out which plugin requires which variable? Hard to pass around, not secure.
4. Plugins manage their own configs themselves.

I prefer number 2, as it's easier to implement and maintain.
