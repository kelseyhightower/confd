# DNS SRV Records

confd can get a list of backend nodes via SRV records. 

## Examples 

### etcd

```
dig SRV _etcd._tcp.confd.io
```

```
... 
;; ANSWER SECTION:
_etcd._tcp.confd.io.  300 IN  SRV 1 50 4001 etcd0.confd.io.
_etcd._tcp.confd.io.  300 IN  SRV 2 50 4001 etcd1.confd.io.
```

```
confd -backend etcd -srv-domain confd.io
```

### consul

```
dig SRV _consul._tcp.confd.io
```

```
...
;; ANSWER SECTION:
_consul._tcp.confd.io.  300 IN  SRV 1 50 8500 consul.confd.io.
```

```
confd -backend consul -srv-domain confd.io
```

```

## The backend scheme

By default the `scheme` will be set to http. If you would like to use https instead use the `-scheme` flag

```Bash
confd -onetime -scheme https -srv-domain confd.io
```

Both the SRV domain and scheme can be configured in the confd configuration file. See the [Configuration Guide](https://github.com/kelseyhightower/confd/wiki/Configuration-Guide) for more details.
