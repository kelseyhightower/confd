# Confd

Manage local configuration files using templates and data from etcd.

## Configuration

```INI
[main]
config_dir = /etc/confd/conf.d/
interval = 30

[etcd]
prefix = /environment/app/uuid
url = http://127.0.0.1:4001
```

## Example 

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
