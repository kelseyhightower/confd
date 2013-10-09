# confd

`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date by polling [etcd](https://github.com/coreos/etcd)
* reloading applications to pick up new config file changes

## Getting Started

### Installing confd

Download the latest binary from [Github](https://github.com/kelseyhightower/confd/releases).

### Building

You can build confd from source:

```
git clone https://github.com/kelseyhightower/confd.git
cd confd
go build
```

This will produce the `confd` binary in the current directory.

## Usage

The following commands will process all the template configs found under `/etc/confd/conf.d`.

### Poll the etcd cluster in 30 second intervals

The "/production" string will be prefixed to keys when querying etcd at http://127.0.0.1:4001.

```
confd -c /etc/confd -i 30 -p '/production' -n 'http://127.0.0.1:4001'
```

### Process template configs once and exit

Using default settings process all template configs and exit.

```
confd -onetime
```

### Client authentication

Same as above but authenticate with client certificates.

```
confd -onetime -key /etc/confd/ssl/client.key -cert /etc/confd/ssl/client.crt
```

## Configuration

The confd configuration file is written in the TOML format and is loaded from
`/etc/confd/confd.toml` by default.

Optional:

* `confdir` (string) - The path to confd configs. The default is /etc/confd.
* `etcd_nodes` (array of strings) - An array of etcd cluster nodes. The default
  is ["http://127.0.0.1:4001"].
* `interval` (int) - The number of seconds to wait between calls to etcd. The
  default is 600.
* `prefix` (string) - The prefix string to prefix to keys when making calls to
  etcd. The default is "/".
* `client_cert` (string) The cert file of the client.
* `client_key` (string) The key file of the client.

Example:

```TOML
[confd]
  confdir = "/etc/confd"
  etcd_nodes = [
    "http://127.0.0.1:4001",
  ]
  interval = 600
  prefix = "/"
  client_cert = "/etc/confd/ssl/client.crt"
  client_key = "/etc/confd/ssl/client.key"
```

## Template Config

Template configs are written in TOML and define a single template resource.
Template configs are stored under the `confdir/conf.d` directory.

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
