# Consul Usage Examples

The following commands will process all the [template resources](https://github.com/kelseyhightower/confd/wiki/Template-Resources) found under `/etc/confd/conf.d`.

> The `-consul` flag is required to use consul as the configuration backend.

### Poll the consul cluster in 30 second intervals

The "production" string will be prefixed to keys when querying consul at http://127.0.0.1:8500.

```Bash
confd -interval 30 -prefix 'production' -consul -consul-addr '127.0.0.1:8500'
```

Note: the prefix will be stripped off key names before they are passed to source templates.

### Same as above in noop mode

```Bash
confd -interval 30 -prefix '/production' -consul -consul-addr '127.0.0.1:8500' -noop
```

See [Noop mode](noop-mode.md)

### Single run without polling

Using default settings run one time and exit.

```
confd -onetime -consul -consul-addr '127.0.0.1:8500'
```

### Client authentication

Current Consul integration does not support client auth. Coming in 0.5.0.

### Lookup consul nodes using SRV records

Current Consul integration does not support SRV records. Coming in 0.5.0.

### Enable verbose logging.

Sometimes you need more details on what's going. Try running confd in verbose mode.

```Bash
confd -consul -verbose
```

You can get even more output with the `-debug` flag

See [Logging Guide](logging.md)
