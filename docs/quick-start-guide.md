# Quick Start Guide

Before we begin be sure to [download and install confd](installation.md). 

## Select a backend

confd support the following backends:

* etcd
* consul
* environment variables

### Add keys

This guide assumes you have a working [etcd](https://github.com/coreos/etcd#getting-started), or [consul](http://www.consul.io/intro/getting-started/install.html) server up and running and the ability to add new keys.

#### etcd

```
etcdctl set /myapp/database/url db.example.com
etcdctl set /myapp/database/user rob
```

#### consul

```
curl -X PUT -d 'db.example.com' http://localhost:8500/v1/kv/myapp/database/url
curl -X PUT -d 'rob' http://localhost:8500/v1/kv/myapp/database/user
```

#### environment variables

```
export MYAPP_DATABASE_URL=db.example.com
export MYAPP_DATABASE_USER=rob
```


### Create the confdir

The confdir is where template resource configs and source templates are stored. The default confdir is `/etc/confd`. Create the confdir by executing the following command:

```Bash
sudo mkdir -p /etc/confd/{conf.d,templates}
```

### Create a template resource config

Template resources are defined in [TOML](https://github.com/mojombo/toml) config files under the `confdir`.

The follow template resource will managed the `/tmp/myconfig.conf` configuration file.

```toml
# /etc/confd/conf.d/myconfig.toml
src = "myconfig.conf.tmpl"
dest = "/tmp/myconfig.conf"
keys = [
    "/myapp/database/url",
    "/myapp/database/user",
]
```

### Create the source template

Source templates are [Golang text templates](http://golang.org/pkg/text/template/#pkg-overview), and are stored under the `confdir` templates directory.

```
# /etc/confd/templates/myconfig.conf.tmpl
[myconfig]
database_url = {{get "/myapp/database/url"}}
database_user = {{get "/myapp/database/user"}}
```

### Processing template resources

confd supports two modes of operation, daemon and onetime mode. In daemon mode, confd runs in the foreground and processing template resources every 5 mins by default. For this tutorial we are going to use onetime mode.

Assuming you etcd server is running at http://127.0.0.1:4001 you can run the following command to process the all the template resources under `/etc/confd/conf.d`:

```
confd -verbose -onetime -node '127.0.0.1:4001'
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
