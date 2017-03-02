# Quick Start Guide

Before we begin be sure to [download and install confd](installation.md).

## Select a backend

confd supports the following backends:

* etcd
* consul
* vault
* environment variables
* redis
* zookeeper
* dynamodb
* stackengine
* rancher
* metad

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

####vault
```
vault mount -path myapp generic
vault write myapp/database url=db.example.com user=rob
```

#### environment variables

```
export MYAPP_DATABASE_URL=db.example.com
export MYAPP_DATABASE_USER=rob
```

#### redis

```
redis-cli set /myapp/database/url db.example.com
redis-cli set /myapp/database/user rob
```

#### zookeeper

```
[zk: localhost:2181(CONNECTED) 1] create /myapp ""
[zk: localhost:2181(CONNECTED) 2] create /myapp/database ""
[zk: localhost:2181(CONNECTED) 3] create /myapp/database/url "db.example.com"
[zk: localhost:2181(CONNECTED) 4] create /myapp/database/user "rob"
```

#### dynamodb

First create a table with the following schema:

```
aws dynamodb create-table \
    --region <YOUR_REGION> --table-name <YOUR_TABLE> \
    --attribute-definitions AttributeName=key,AttributeType=S \
    --key-schema AttributeName=key,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1
```

Now create the items. The attribute value `value` must be of type [string](http://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html):

```
aws dynamodb put-item --table-name <YOUR_TABLE> --region <YOUR_REGION> \
    --item '{ "key": { "S": "/myapp/database/url" }, "value": {"S": "db.example.com"}}'
aws dynamodb put-item --table-name <YOUR_TABLE> --region <YOUR_REGION> \
    --item '{ "key": { "S": "/myapp/database/user" }, "value": {"S": "rob"}}'
```
#### StackEngine
```
curl -k -X PUT -d 'value' https://mesh-01:8443/api/kv/key --header "Authorization: Bearer stackengine_api_key"
```

#### Rancher

This backend consumes the [Rancher](https://www.rancher.com) metadata service. For available keys, see the [Rancher Metadata Service docs](http://docs.rancher.com/rancher/rancher-services/metadata-service/).

#### Metad
```
curl http://127.0.0.1:9611/v1/data -X PUT -d '{"myapp":{"database":{"url":"db.example.com","user":"rob"}}}'
```

### Create the confdir

The confdir is where template resource configs and source templates are stored.

```
sudo mkdir -p /etc/confd/{conf.d,templates}
```

### Create a template resource config

Template resources are defined in [TOML](https://github.com/mojombo/toml) config files under the `confdir`.

/etc/confd/conf.d/myconfig.toml
```
[template]
src = "myconfig.conf.tmpl"
dest = "/tmp/myconfig.conf"
keys = [
    "/myapp/database/url",
    "/myapp/database/user",
]
```

### Create the source template

Source templates are [Golang text templates](http://golang.org/pkg/text/template/#pkg-overview).

/etc/confd/templates/myconfig.conf.tmpl
```
[myconfig]
database_url = {{getv "/myapp/database/url"}}
database_user = {{getv "/myapp/database/user"}}
```

### Process the template

confd supports two modes of operation daemon and onetime. In daemon mode confd polls a backend for changes and updates destination configuration files if necessary.

#### etcd

```
confd -onetime -backend etcd -node http://127.0.0.1:4001
```

#### consul

```
confd -onetime -backend consul -node 127.0.0.1:8500
```

#### vault
```
ROOT_TOKEN=$(vault read -field id auth/token/lookup-self)

confd -onetime -backend vault -node http://127.0.0.1:8200 \
      -auth-type token -auth-token $ROOT_TOKEN
```

#### dynamodb

```
confd -onetime -backend dynamodb -table <YOUR_TABLE>
```

#### env

```
confd -onetime -backend env
```

#### StackEngine

```
confd -onetime -backend stackengine -auth-token stackengine_api_key -node 192.168.255.210:8443 -scheme https
```

#### redis

```
confd -onetime -backend redis -node 192.168.255.210:6379
```

#### rancher

```
confd -onetime -backend rancher -prefix /2015-07-25
```

#### metad

```
confd --onetime --backend metad --node 127.0.0.1:80 --watch
```
=======

*Note*: The metadata api prefix can be defined on the cli, or as part of your keys in the template toml file.

Output:
```
2014-07-08T20:38:36-07:00 confd[16252]: INFO Target config /tmp/myconfig.conf out of sync
2014-07-08T20:38:36-07:00 confd[16252]: INFO Target config /tmp/myconfig.conf has been updated
```

The `dest` configuration file should now be in sync.

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

## Advanced Example

In this example we will use confd to manage two nginx config files using a single template.

### Add keys

#### etcd

```
etcdctl set /myapp/subdomain myapp
etcdctl set /myapp/upstream/app2 "10.0.1.100:80"
etcdctl set /myapp/upstream/app1 "10.0.1.101:80"
etcdctl set /yourapp/subdomain yourapp
etcdctl set /yourapp/upstream/app2 "10.0.1.102:80"
etcdctl set /yourapp/upstream/app1 "10.0.1.103:80"
```

#### consul

```
curl -X PUT -d 'myapp' http://localhost:8500/v1/kv/myapp/subdomain
curl -X PUT -d '10.0.1.100:80' http://localhost:8500/v1/kv/myapp/upstream/app1
curl -X PUT -d '10.0.1.101:80' http://localhost:8500/v1/kv/myapp/upstream/app2
curl -X PUT -d 'yourapp' http://localhost:8500/v1/kv/yourapp/subdomain
curl -X PUT -d '10.0.1.102:80' http://localhost:8500/v1/kv/yourapp/upstream/app1
curl -X PUT -d '10.0.1.103:80' http://localhost:8500/v1/kv/yourapp/upstream/app2
```

### Create template resources

/etc/confd/conf.d/myapp-nginx.toml

```
[template]
prefix = "/myapp"
src = "nginx.tmpl"
dest = "/tmp/myapp.conf"
owner = "nginx"
mode = "0644"
keys = [
  "/subdomain",
  "/upstream",
]
check_cmd = "/usr/sbin/nginx -t -c {{.src}}"
reload_cmd = "/usr/sbin/service nginx reload"
```

/etc/confd/conf.d/yourapp-nginx.toml

```
[template]
prefix = "/yourapp"
src = "nginx.tmpl"
dest = "/tmp/yourapp.conf"
owner = "nginx"
mode = "0644"
keys = [
  "/subdomain",
  "/upstream",
]
check_cmd = "/usr/sbin/nginx -t -c {{.src}}"
reload_cmd = "/usr/sbin/service nginx reload"
```

### Create the source template

/etc/confd/templates/nginx.tmpl
```
upstream {{getv "/subdomain"}} {
{{range getvs "/upstream/*"}}
    server {{.}};
{{end}}
}

server {
    server_name  {{getv "/subdomain"}}.example.com;
    location / {
        proxy_pass        http://{{getv "/subdomain"}};
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
   }
}
```
