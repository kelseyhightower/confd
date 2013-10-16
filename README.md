# confd

[![Build Status](https://travis-ci.org/kelseyhightower/confd.png?branch=master)](https://travis-ci.org/kelseyhightower/confd)

`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date by polling [etcd](https://github.com/coreos/etcd) and processing [template resources](https://github.com/kelseyhightower/confd#template-resources).
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

The following commands will process all the [template resources](https://github.com/kelseyhightower/confd#template-resources) found under `/etc/confd/conf.d`.

### Poll the etcd cluster in 30 second intervals

The "/production" string will be prefixed to keys when querying etcd at http://127.0.0.1:4001.

```
confd -c /etc/confd -i 30 -p '/production' -n 'http://127.0.0.1:4001'
```

### Single run without polling

Using default settings run one time and exit.

```
confd -onetime
```

### Client authentication

Same as above but authenticate with client certificates.

```
confd -onetime -key /etc/confd/ssl/client.key -cert /etc/confd/ssl/client.crt
```

## Configuration

The confd configuration file is written in [TOML](https://github.com/mojombo/toml)
and loaded from `/etc/confd/confd.toml` by default.

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
confdir  = "/etc/confd"
interval = 600
prefix   = "/"
etcd_nodes = [
  "http://127.0.0.1:4001",
]
client_cert = "/etc/confd/ssl/client.crt"
client_key  = "/etc/confd/ssl/client.key"
```

## Template Resources

Template resources are written in TOML and define a single template resource.
Template resources are stored under the `confdir/conf.d` directory.

Required:

* `dest` (string) - output file where the template should be rendered.
* `keys` (array of strings) - An array of etcd keys. Keys will be looked up
  with the configured prefix.
* `src` (string) - relative path of a the [configuration template](https://github.com/kelseyhightower/confd#templates).

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
src   = "nginx.conf.tmpl"
dest  = "/etc/nginx/nginx.conf"
owner = "root"
group = "root"
mode  = "0644"
keys = [
  "/nginx",
]
check_cmd  = "/usr/sbin/nginx -t -c {{ .src }}"
reload_cmd = "/usr/sbin/service nginx restart"
```

## Templates

Templates define a single application configration template.
Templates are stored under the `confdir/templates` directory.

Templates are written in Go's [`text/template`](http://golang.org/pkg/text/template/). 

Etcd keys are treated as paths and automatically transformed into keys for retrieval in templates. Underscores are used in place of forward slashes.  _Values retrived from Etcd are never modified._  
For example `/foo/bar` becomes `foo_bar`.

`foo_bar` is accessed as `{{ .foo_bar }}`


Example:  
```
$ etcdctl set /nginx/domain 'example.com'
$ etcdctl set /nginx/root '/var/www/example_dotcom'
$ etcdctl set /nginx/worker_processes '2'
```


`$ cat /etc/confd/templates/nginx.conf.tmpl`:
```
worker_processes {{ .nginx_worker_processes }};

server {
    listen 80;
    server_name www.{{ .nginx_domain }};
    access_log /var/log/nginx/{{ .nginx_domain }}.access.log;
    error_log /var/log/nginx/{{ .nginx_domain }}.log;

    location / {
        root   {{ .nginx_root }};
        index  index.html index.htm;
    }
}
```

Will produce `/etc/nginx/nginx.conf`:
```
worker_processes 2;

server {
    listen 80;
    server_name www.example.com;
    access_log /var/log/nginx/example.com.access.log;
    error_log /var/log/nginx/example.com.error.log;

    location / {
        root   /var/www/example_dotcom;
        index  index.html index.htm;
    }
}
```

Go's [`text/template`](http://golang.org/pkg/text/template/) package is very powerful. For more details on it's capabilities see its [documentation.](http://golang.org/pkg/text/template/)
