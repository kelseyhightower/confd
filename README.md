# confd

`confd` is a lightweight configuration management tool focused on keeping local
configuration files up-to-date by polling [etcd](https://github.com/coreos/etcd)
for specific keys and regenerating templates when values change. `confd` can also
take care of reloading applications to pick up new config file changes.

## Install

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
  "services": {
    "example": {
      "reload_cmd": "/usr/bin/touch /tmp/example-reloaded"
    }
  }
}
```

## Configuration

```INI
[main]
config_dir = /etc/confd/conf.d/
interval = 30

[etcd]
prefix = /environment/app/uuid
url = http://127.0.0.1:4001
```
