# Configuration Guide

The confd configuration file is written in [TOML](https://github.com/mojombo/toml)
and loaded from `/etc/confd/confd.toml` by default. You can specify the config file via the `-config-file` command line flag.

> Note: You can use confd without a configuration file. See [Command Line Flags](https://github.com/kelseyhightower/confd/wiki/Command-Line-Flags).

Optional:

* `backend` (string) - The backend to use. ("etcd")
* `client_cakeys` (string) - The client CA key file.
* `client_cert` (string) - The client cert file.
* `client_key` (string) - The client key file.
* `confdir` (string) - The path to confd configs. ("/etc/confd")
* `debug` (bool) - Enable debug logging.
* `interval` (int) - The backend polling interval. (600)
* `nodes` (array of strings) - List of backend nodes. (["http://127.0.0.1:4001"])
* `noop` (bool) - Enable noop mode. Process all template resources; skip target update.
* `prefix` (string) - The string to prefix to keys. ("/")
* `quiet` (bool) - Enable quiet logging.
* `scheme` (string) - The backend URI scheme. ("http" or "https")
* `srv_domain` (string) - The name of the resource record.
* `verbose` (bool) - Enable verbose logging.

Example:

```TOML
backend = "etcd"
client_cert = "/etc/confd/ssl/client.crt"
client_key = "/etc/confd/ssl/client.key"
confdir = "/etc/confd"
debug = false
interval = 600
nodes = [
  "http://127.0.0.1:4001",
]
noop = false
prefix = "/production"
quiet = false
scheme = "https"
srv_domain = "etcd.example.com"
verbose = false
```
