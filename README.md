# confd

`confd` is a lightweight configuration management tool focused on keeping local
configuration files up-to-date by polling [etcd](https://github.com/coreos/etcd)
for specific keys and regenerating templates when values change. `confd` can also
take care of reloading applications to pick up new config file changes.

## Install

```
go get github.com/kelseyhightower/confd
```

## Configuration

confd loads external configuration from `/etc/confd/confd.toml`

```TOML
configdir = "/etc/confd/conf.d"
interval = 600
prefix = "/production/app"
etcd_nodes = [
  "http://127.0.0.1:4001",
  "http://127.0.0.1:4002"
]
```

## confd Configs

`confd` configs are written in TOML and define a single template resource.
`confd` configs are stored under the `confdir` directory.

Example:

```TOML
keys = [
  "/nginx/port",
  "/nginx/servername"
]
src = "nginx.conf.tmpl"
dest = "/etc/nginx/nginx.conf"
owner = "root"
group = "root"
mode = "0644"
reload_cmd = "/sbin/service nginx reload"
```

## Template Resource

Required:

 * `dest` - output file where the template should be rendered.
 * `keys` - list of etcd keys. Keys will be looked up with the configured prefix.
 * `src` - relative path of a Go template.

Optional:

 * `group` - name of the group that should own the file.
 * `mode` - mode the file should be in.
 * `owner` - name of the user that should own the file.
 * `reload_cmd` - command to reload config.

