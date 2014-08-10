# Templates

Templates define a single application configration template.
Templates are stored under the `/etc/confd/templates` directory by default.

Templates are written in Go's [`text/template`](http://golang.org/pkg/text/template/). 

## Template Functions

### base

Alias for the path.Base function.

```
{{with get "/key"}}
    key: {{base .Key}}
    value: {{.Value}}
{{end}}
```

### get

Returns the KVPair where key matches its argument. Returns an error if key is not found.

```
{{with get "/key"}}
    key: {{.Key}}
    value: {{.Value}}
{{end}}
```

### gets

Returns all KVPair, []KVPair, where key matches its argument. Returns an error if key is not found.

```
{{range gets "/*"}}
    key: {{.Key}}
    value: {{.Value}}
{{end}}
```

### getv

Returns the value as a string where key matches its argument. Returns an error if key is not found.

```
value: {{getv "/key"}}
```

### getvs

Returns all values, []string, where key matches its argument. Returns an error if key is not found.

```
{{range getvs "/*"}}
    value: {{.}}
{{end}}
```

### ls

Returns all subkeys, []string, where path matches its argument. Returns an empty list if path is not found.

```
{{range ls "/deis/services"}}
   value: {{.}}
{{end}}
```

### lsdir

Returns all subkeys, []string, where path matches its argument. It only returns subkeys that also have subkeys. Returns an empty list if path is not found.

```
{{range lsdir "/deis/services"}}
   value: {{.}}
{{end}}
```

## Example Usage  

```Bash
etcdctl set /nginx/domain 'example.com'
etcdctl set /nginx/root '/var/www/example_dotcom'
etcdctl set /nginx/worker_processes '2'
etcdctl set /app/upstream/app1 "10.0.1.100:80"
etcdctl set /app/upstream/app2 "10.0.1.101:80"
```

`/etc/confd/templates/nginx.conf.tmpl`

```Text
worker_processes {{getv "/nginx/worker_processes"}};

upstream app {
{{range getvs "/app/upstream/*"}}
    server {{.}};
{{end}}
}

server {
    listen 80;
    server_name www.{{getv "/nginx/domain"}};
    access_log /var/log/nginx/{{getv "/nginx/domain"}}.access.log;
    error_log /var/log/nginx/{{getv "/nginx/domain"}}.log;

    location / {
        root              {{getv "/nginx/root"}};
        index             index.html index.htm;
		proxy_pass        http://app;
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
    }
}
```

Output: `/etc/nginx/nginx.conf`

```Text
worker_processes 2;

upstream app {
    server 10.0.1.100:80;
    server 10.0.1.101:80;
}

server {
    listen 80;
    server_name www.example.com;
    access_log /var/log/nginx/example.com.access.log;
    error_log /var/log/nginx/example.com.error.log;

    location / {
        root              /var/www/example_dotcom;
        index             index.html index.htm;
        proxy_pass        http://app;
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
    }
}
```

Go's [`text/template`](http://golang.org/pkg/text/template/) package is very powerful. For more details on it's capabilities see its [documentation.](http://golang.org/pkg/text/template/)
