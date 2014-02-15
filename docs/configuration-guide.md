# Configuration Guide

The confd configuration file is written in [TOML](https://github.com/mojombo/toml)
and loaded from `/etc/confd/confd.toml` by default. You can specify a different config file location via the `-config-file` command line flag.

```Bash
confd -config-file ~/confd.toml
```

> Note: You can use confd without a configuration file. See [Command Line Flags](https://github.com/kelseyhightower/confd/wiki/Command-Line-Flags).

Optional:

* `debug` (bool) - Enable debug logging.
* `client_cert` (string) The cert file of the client.
* `client_key` (string) The key file of the client.
* `confdir` (string) - The path to confd configs. The default is /etc/confd.
* `etcd_nodes` (array of strings) - An array of etcd cluster nodes. The default
  is ["http://127.0.0.1:4001"].
* `etcd_scheme` (string) - The etcd scheme to use. Must be 'http' or 'https'
* `interval` (int) - The number of seconds to wait between calls to etcd. The
  default is 600.
* `noop` (bool) - Enable noop mode. Process all template resource, but don't update target config.
* `prefix` (string) - The prefix string to prefix to keys when making calls to
  etcd. The default is "/".
* `quiet` (bool) - Enable quiet logging. Only error messages are printed.
* `srv_domain` (string) - The domain to query for etcd SRV records.
* `verbose` (bool) - Enable verbose logging.

Example:

```TOML
[confd]
confdir  = "/etc/confd"
interval = 600
prefix   = "/"
etcd_nodes = [
  "http://127.0.0.1:4001",
]
client_cert = "/etc/confd/ssl/client.crt"
client_key  = "/etc/confd/ssl/client.key"
```
