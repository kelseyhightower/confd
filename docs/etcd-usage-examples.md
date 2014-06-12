# Usage Examples

The following commands will process all the [template resources](https://github.com/kelseyhightower/confd/wiki/Template-Resources) found under `/etc/confd/conf.d`.

### Poll the etcd cluster in 30 second intervals

The "/production" string will be prefixed to keys when querying etcd at http://127.0.0.1:4001.

```Bash
confd -interval 30 -prefix '/production' -node 'http://127.0.0.1:4001'
```

Note: the prefix will be stripped off key names before they are passed to source templates.

### Same as above in noop mode

```Bash
confd -interval 30 -prefix '/production' -node 'http://127.0.0.1:4001' -noop
```

See [Noop mode](noop-mode.md)

### Single run without polling

Using default settings run one time and exit.

```
confd -onetime
```

### Client authentication

Same as above but authenticate with client certificates.

```
confd -onetime -client-key /etc/confd/ssl/client.key -client-cert /etc/confd/ssl/client.crt
```

### Lookup etcd nodes using SRV records

```
dig SRV _etcd._tcp.confd.io
...
;; ANSWER SECTION:
_etcd._tcp.confd.io.  300 IN  SRV 1 50 4001 etcd0.confd.io.
_etcd._tcp.confd.io.  300 IN  SRV 2 50 4001 etcd1.confd.io.
```

```
confd -srv-domain example.com -etcd-scheme https
```

confd would connect to the nodes at `["https://etcd0.confd.io:4001", "https://etcd1.confd.io:4001"]`

See [Using etcd SRV Records](dns-srv-records.md)

### Enable verbose logging.

Sometimes you need more details on what's going. Try running confd in verbose mode.

```Bash
confd -verbose
```

You can get even more output with the `-debug` flag

See [Logging Guide](logging.md)
