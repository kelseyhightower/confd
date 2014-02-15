# Template Resources

Template resources are written in TOML and define a single template resource.
Template resources are stored under the `/etc/confd/conf.d` directory by default.

Required:

* `dest` (string) - output file where the template should be rendered.
* `keys` (array of strings) - An array of etcd keys. Keys will be looked up
  with the configured prefix.
* `src` (string) - relative path of a [configuration template](templates.md).

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
