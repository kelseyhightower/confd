# Template Resources

Template resources are written in TOML and define a single template resource.
Template resources are stored under the `/etc/confd/conf.d` directory by default.

### Required

* `dest` (string) - The target file.
* `keys` (array of strings) - An array of keys.
* `src` (string) - The relative path of a [configuration template](templates.md).

### Optional

* `gid` (int) - The gid that should own the file. Defaults to the effective gid.
* `mode` (string) - The permission mode of the file.
* `uid` (int) - The uid that should own the file. Defaults to the effective uid.
* `reload_cmd` (string) - The command to reload config.
* `check_cmd` (string) - The command to check config. Use `{{.src}}` to reference the rendered source template.
* `prefix` (string) - The string to prefix to keys.

### Notes

When using the `reload_cmd` feature it's important that the command exits on its own. The reload
command is not managed by confd, and will block the configuration run until it exits.

## Example

```TOML
[template]
src = "nginx.conf.tmpl"
dest = "/etc/nginx/nginx.conf"
uid = 0
gid = 0
mode = "0644"
keys = [
  "/nginx",
]
check_cmd = "/usr/sbin/nginx -t -c {{.src}}"
reload_cmd = "/usr/sbin/service nginx restart"
```
