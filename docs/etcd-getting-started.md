# Quick Start with etcd

### Add keys to etcd

This guide assumes you have a working [etcd](https://github.com/coreos/etcd#getting-started) server up and running and the ability to add new keys. Using the `etcdctl` command line tool add the following keys and values to etcd:

```
etcdctl set /myapp/database/url db.example.com
etcdctl set /myapp/database/user rob
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

### Processing template resources

confd supports two modes of operation, daemon and onetime mode. In daemon mode, confd runs in the foreground processing template resources every 5 mins by default. For this tutorial we are going to use onetime mode.

Assuming your etcd server is running at http://127.0.0.1:4001 you can run the following command to process the `~/confd/conf.d/myconfig.toml` template resource:

```
confd -verbose -onetime -node 'http://127.0.0.1:4001' -confdir ~/confd
```
Output:
```
2013-11-03T18:00:47-08:00 confd[21294]: NOTICE Starting confd
2013-11-03T18:00:47-08:00 confd[21294]: NOTICE etcd nodes set to http://127.0.0.1:4001
2013-11-03T18:00:47-08:00 confd[21294]: INFO Target config /tmp/myconfig.conf out of sync
2013-11-03T18:00:47-08:00 confd[21294]: INFO Target config /tmp/myconfig.conf has been updated
```

The `dest` config should now be in sync with the template resource configuration.

```
cat /tmp/myconfig.conf
```

Output:
```
# This a comment
[myconfig]
database_url = db.example.com
database_user = rob
```

## Next steps

Check out the [docs directory](.) for more docs and [usage examples](etcd-usage-examples.md).
