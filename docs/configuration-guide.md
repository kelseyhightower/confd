# Configuration Guide

The confd configuration file is written in [TOML](https://github.com/mojombo/toml)
and loaded from `/etc/confd/confd.toml` by default. You can specify the config file via the `-config-file` command line flag.

> Notes:
> - You can use confd without a configuration file. See [Command Line Flags](https://github.com/kelseyhightower/confd/blob/master/docs/command-line-flags.md).
> - The confdir cannot be a symlink due to lack of support in underlying functions. This was discussed in issue #741

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
* `auth_token` (string) - Auth bearer token to use.
* `auth_type` (string) - Vault auth backend type to use.
* `basic_auth` (bool) - Use Basic Auth to authenticate (only used with -backend=consul and -backend=etcd).
* `table` (string) - The name of the DynamoDB table (only used with -backend=dynamodb).
* `separator` (string) - The separator to replace '/' with when looking up keys in the backend, prefixed '/' will also be removed (only used with -backend=redis)
* `username` (string) - The username to authenticate as (only used with vault and etcd backends).
* `password` (string) - The password to authenticate with (only used with vault and etcd backends).
* `app_id` (string) - Vault app-id to use with the app-id backend (only used with -backend=vault and auth-type=app-id).
* `user_id` (string) - Vault user-id to use with the app-id backend (only used with -backend=value and auth-type=app-id).
* `role_id` (string) - Vault role-id to use with the AppRole, Kubernetes backends (only used with -backend=vault and either auth-type=app-role or auth-type=kubernetes).
* `secret_id` (string) - Vault secret-id to use with the AppRole backend (only used with -backend=vault and auth-type=app-role).
* `file` (array of strings) - The YAML file to watch for changes (only used with -backend=file).
* `filter` (string) - Files filter (only used with -backend=file) (default "*").
* `path` (string) - Vault mount path of the auth method (only used with -backend=vault).

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
