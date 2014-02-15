# Templates

Templates define a single application configration template.
Templates are stored under the `/etc/confd/templates` directory by default.

Templates are written in Go's [`text/template`](http://golang.org/pkg/text/template/). 

> Etcd keys are treated as paths and automatically transformed into keys for retrieval in templates. Underscores are used in place of forward slashes.  _Values retrived from Etcd are never modified._  
> For example `/foo/bar` becomes `foo_bar`.
> `foo_bar` is accessed as `{{ .foo_bar }}`


Example:  
```Bash
$ etcdctl set /nginx/domain 'example.com'
$ etcdctl set /nginx/root '/var/www/example_dotcom'
$ etcdctl set /nginx/worker_processes '2'
```


`$ cat /etc/confd/templates/nginx.conf.tmpl`:
```Text
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
```Text
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
