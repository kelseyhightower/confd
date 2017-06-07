# Plugin system draft

Plugin system is implemented using https://github.com/hashicorp/go-plugin, which is successfully used in Terraform, Packer, Nomad and Vault. It's based on network RPC over local network. Plugins are Go interface implementations, which maps quite good to the current backends implementation.

The core set of plugins is defined in builtin/ directory. They are compiled in the main binary.

Builtin plugins are called using special cli command:
`confd internal-plugin <pluginType> <pluginName>``

## Open questions
Each plugin should have it's own config.
Confd should only know which plugin to use, but shouldn't know about any configuration options.

How plugins should be configured?
1. Plugins should be configured per template. Less reusable. Easy to understand. Hard to configure complex backends, which tend to have a lot of configuration options.
2. Plugins should be configured globally via a file, e.g. per plugin ~/.confd/etcd.toml or all-in-one ~/.confd.toml. Less repeatable then 1,
3. Plugins should be configured via environment variables. How to figure our which plugin requires which variable? Hard to pass around, not secure.
4. Plugins manage their own configs themselves.

I prefer number 2 for now, as it's easier to implement and maintain. But it can be changed later on.
