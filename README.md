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

The confd configuration file is written in the TOML format and is loaded from
`/etc/confd/confd.toml` by default.

Optional:

 * `confdir` (string) - The path to confd configs. The default is "/etc/confd".
 * `etcd_nodes` (array of strings) - An array of etcd cluster nodes. The default
   is ["http://127.0.0.1:4001"].
 * `interval` (int) - The number of seconds wait between calls to etcd. The
   default is 600.
 * `prefix` (string) - The prefix string to prefix to keys when making calls to
   etcd. The default is "/".

Example:

```TOML
[confd]
  confdir = "/etc/confd"
  etcd_nodes = [
    "http://127.0.0.1:4001",
  ]
  interval = 600
  prefix = "/"
```

## confd Configs

`confd` configs are written in TOML and define a single template resource.
`confd` configs are stored under the `confdir/conf.d` directory.

Example:

```TOML
[template]
  src = "nginx.conf.tmpl"
  dest = "/etc/nginx/nginx.conf"
  group = "root"
  keys = [
    "/nginx/worker_processes",
  ]
  owner = "root"
  mode = "0644"
  check_cmd = "/usr/sbin/nginx -t -c {{ .src }}"
  reload_cmd = "/usr/sbin/service nginx restart"
```

### Template Resource

Required:

 * `dest` (string) - output file where the template should be rendered.
 * `keys` (array of strings) - An array of etcd keys. Keys will be looked up
   with the configured prefix.
 * `src` (string) - relative path of a Go template. Templates are stored under
   the `confdir/templates` directory.

Optional:

 * `group` (string) - name of the group that should own the file.
 * `mode` (string) - mode the file should be in.
 * `owner` (string) - name of the user that should own the file.
 * `reload_cmd` (string) - command to reload config.
 * `check_cmd` (string) - command to check config. Use `{{ .src }}` to reference
   the rendered source template.

