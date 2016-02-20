# Configuration Guide

The confd configuration file is written in [TOML](https://github.com/mojombo/toml)
and loaded from `/etc/confd/confd.toml` by default. You can specify the config file via the `-config-file` command line flag.

> Note: You can use confd without a configuration file. See [Command Line Flags](https://github.com/kelseyhightower/confd/blob/master/docs/command-line-flags.md).

Optional:

* `backend` (string) - The backend to use. ("etcd")
* `client_cakeys` (string) - The client CA key file.
* `client_cert` (string) - The client cert file.
* `client_key` (string) - The client key file.
* `confdir` (string) - The path to confd configs. ("/etc/confd")
* `interval` (int) - The backend polling interval in seconds. (600)
* `log-level` (string) - level which confd should log messages ("info")
* `nodes` (array of strings) - List of backend nodes. (["http://127.0.0.1:4001"])
* `noop` (bool) - Enable noop mode. Process all template resources; skip target update.
* `prefix` (string) - The string to prefix to keys. ("/")
* `scheme` (string) - The backend URI scheme. ("http" or "https")
* `srv_domain` (string) - The name of the resource record.
* `srv_record` (string) - The SRV record to search for backends nodes.
* `sync-only` (bool) - sync without check_cmd and reload_cmd.
* `watch` (bool) - Enable watch support.

Example:

```TOML
backend = "etcd"
client_cert = "/etc/confd/ssl/client.crt"
client_key = "/etc/confd/ssl/client.key"
confdir = "/etc/confd"
log-level = "debug"
interval = 600
nodes = [
  "http://127.0.0.1:4001",
]
noop = false
prefix = "/production"
scheme = "https"
srv_domain = "etcd.example.com"
```
