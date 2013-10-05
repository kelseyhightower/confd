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
[confd]
configdir = "/etc/confd/conf.d"
templatedir = "/etc/confd/templates"
interval = 600

[etcd]
prefix = "/production/app"
machines = ["http://127.0.0.1:4001", "http://127.0.0.1:4002"]
```

## Config Templates

Config templates are used to define a group of configuration file resources and
a set of etcd keys that can be used to generate application configuration files.
Config templates also define reload commands that should be run to notify or
force an application to pick up changes.

Config templates are written in the JSON format and stored under the
`/etc/confd/conf.d/` directory.

## Example 

`/etc/confd/conf.d/example.json`

```JSON
{
  "templates": [
    {
      "keys": [
        "/example/port",
        "/example/password"
      ],
      "src": "example.conf.tmpl",
      "dest": "/etc/example.conf",
      "owner": "root",
      "group": "root",
      "mode": "0644",
      "service": "example"
    }
  ],
  "services": [
    {
      "name": "example",
      "cmd": "/usr/bin/touch /tmp/example-reloaded"
    }
  ]
}
```

## Resources

### Template Resource

Templates are processed by the Go `text/template` package.

Required:

 * `dest` - output file where the template should be rendered.
 * `keys` - list of etcd keys. Keys will be looked up with the configured prefix.
 * `src` - relative path of a Go template.

Optional:

 * `group` - name of the group that should own the file.
 * `mode` - mode the file should be in.
 * `owner` - name of the user that should own the file.
 * `service` - name of the service resource that should be notified on changes.

Example:

 * `/etc/confd/conf.d/nginx.json`
 * `/etc/confd/templates/nginx.conf.tmpl`

```JSON
{
  "keys": [
    "/nginx/port",
    "/nginx/servername"
  ],
  "src": "nginx.conf.tmpl",
  "dest": "/etc/nginx/nginx.conf",
  "owner": "root",
  "group": "root",
  "mode": "0644",
  "service": "nginx"
}
```

### Service Resource

Required:

 * `name` - name of the service.
 * `cmd` - command that should be executed on changes.
