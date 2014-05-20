# Quick Start with consul

### Add keys to consul

This guide assumes you have a working [consul](http://www.consul.io/intro/getting-started/install.html) server up and running and the ability to add new keys. Using the `curl` command line tool add the following keys and values to consul:

```
curl -X PUT -d 'db.example.com' http://localhost:8500/v1/kv/myapp/database/url
curl -X PUT -d 'rob' http://localhost:8500/v1/kv/myapp/database/user
```

### Create the confdir

The confdir is where template resource configs and source templates are stored. The default confdir is `/etc/confd`. Create the confdir by executing the following command:

```Bash
sudo mkdir -p /etc/confd/{conf.d,templates}
```

You don't have to use the default `confdir` location. For example you can create the confdir under your home directory. Then you tell confd to use the new `confdir` via the `-confdir` flag.

```Bash
mkdir -p ~/confd/{conf.d,templates}
```

### Create a template resource config

Template resources are defined in [TOML](https://github.com/mojombo/toml) config files under the `confdir` conf.d directory (i.e. ~/confd/conf.d/*.toml).

Create the following template resource config and save it as `~/confd/conf.d/myconfig.toml`.

```Text
[template]
src = "myconfig.conf.tmpl"
dest = "/tmp/myconfig.conf"
keys = [
  "/myapp/database/url",
  "/myapp/database/user",
]
```

### Create the source template

Source templates are plain old [Golang text templates](http://golang.org/pkg/text/template/#pkg-overview), and are stored under the `confdir` templates directory. Create the following source template and save it as `~/confd/templates/myconfig.conf.tmpl`

```
# This a comment
[myconfig]
database_url = {{ .myapp_database_url }}
database_user = {{ .myapp_database_user }}
```

### Run confd

```
confd -onetime -consul -consul-addr localhost:8500
```

## Next steps

Checkout the [docs directory](docs) for more docs and [usage examples](consul-usage-examples.md).
