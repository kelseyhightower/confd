# Templates - Interation Example

Using confd to manage nginx proxy config

## Add upstream servers to etcd

```Bash
curl http://127.0.0.1:4001/v2/keys/myapp/upstream -XPUT -d dir=true
curl http://127.0.0.1:4001/v2/keys/myapp/upstream/app2 -XPUT -d value="10.0.1.101:80"
curl http://127.0.0.1:4001/v2/keys/myapp/upstream/app1 -XPUT -d value="10.0.1.100:80"
```

## Create a template resource

`/etc/confd/conf.d/myapp-nginx.toml`

```TOML
[template]
keys = [
  "myapp/upstream",
]
owner = "nginx"
mode = "0644"
src = "myapp-nginx.tmpl"
dest = "/tmp/myapp.conf"
check_cmd  = "/usr/sbin/nginx -t -c {{ .src }}"
reload_cmd = "/usr/sbin/service nginx reload"
```

## Create a source template

`/etc/confd/templates/nginx.tmpl`

```
upstream myapp {
{{range $server := .myapp_upstream}}
    server {{$server.Value}};
{{end}}
}
 
server {
    server_name  www.example.com;
 
    location / {
        proxy_pass        http://myapp;
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
   }
}
```

## Run confd

```
confd -onetime
2014-02-16T21:36:46-08:00 confd[2180]: INFO Target config /tmp/nginx.conf out of sync
2014-02-16T21:36:46-08:00 confd[2180]: INFO Target config /tmp/nginx.conf has been updated
```

Output

```
upstream myapp {

    server 10.0.1.100:80;

    server 10.0.1.101:80;

}
 
server {
    server_name  www.example.com;
 
    location / {
        proxy_pass        http://myapp;
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
   }
}
```
