# Plugin system

The plugin system is implemented using https://github.com/hashicorp/go-plugin, which is successfully used in Terraform, Packer, Nomad and Vault. It's based on either gRPC or net/rpc over the local network connection. Plugins are Go interface implementations, which maps quite good to the current backends implementation.

The core set of plugins is defined in `builtin/` directory. They are compiled in the main binary.

Builtin plugins are called using the special cli command:
```sh
confd internal-plugin <pluginType> <pluginName>
```

The only `pluginType` in use is `database`.

Plugin discovery mechanism will look for binaries named using `confd-database-*` format in the following locations:
1. A path where confd is installed
2. A path where confd is invoked

To preserve backward compatibility, configuration is passed to builtin plugins in the form of a `map[string]string` as a parameter to the `Configure` method.

A third-party plugin should manage it's own config directly. The `map[string]string` passed to a third-party plugin in `Configure` method will be empty.

## Interface changes

The new interface which plugins implement has the new method `Configure`, which is used to decouple plugin configuration from a core of confd. This method is called to initialize a plugin with a config. A config is passed to a plugin in a form of `map[string]string`. Config deserialization is done using `github.com/mitchellh/mapstructure`, which is a simple and a safe way to convert a map to a Go struct. A plugin may extract values from a map directly, but this behavior is unsafe and discouraged.

`stopChan` parameter was removed from `WatchPrefix` method, because it wasn't necessary and is hard to serialize correctly.

```go
type Database interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64) (uint64, error)
	Configure(map[string]string) error
}
```

## How to write a plugin

Any builtin plugin can be used as an example on how to write a plugin for confd.
All builtin plugins can be compiled to external plugins by running `go build` from the respectful plugin directory in `builtin/bins`. For example, to compile env plugin to an external plugin:
```sh
cd builtin/bins/env
go build -o ../../../bin/confd-database-env-test
```

The code which is used to run a plugin is very simple:
```go
package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/env"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		DatabaseFunc: env.Database,
	})
}
```

You can use generated `confd-database-env-test` plugin binary like that:
```sh
./bin/confd --confdir ./integration/confdir --backend env-test --one-time
```
