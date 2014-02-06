# DNS SRV Records

By default confd will try and connect to etcd at `http://127.0.0.1:4001`. To connect to a different set of etcd nodes you can use the `-node` flag on the command line:

```Bash
confd -node http://etcd0.example.com:4001 -node http://etcd1.example.com:4001
```

Or via the `etcd_nodes` configuration option in the confd configuration file, i.e. `/etc/confd/confd.toml`.

```TOML
etcd_nodes = [
  "http://etcd0.example.com:4001",
  "http://etcd1.example.com:4001",
]
```

As of confd version 0.2.0 you can specify a DNS domain name to query for etcd SRV records via the `-srv-domain` flag. Assuming the following SRV records exist:

```Bash
 dig SRV _etcd._tcp.confd.io
```

```Bash
... 
;; ANSWER SECTION:
_etcd._tcp.confd.io.  300 IN  SRV 1 50 4001 etcd0.confd.io.
_etcd._tcp.confd.io.  300 IN  SRV 2 50 4001 etcd1.confd.io.
```

```Bash
confd -onetime -verbose -srv-domain confd.io
```

Output
```Bash
2013-11-03T19:04:53-08:00 confd[21356]: INFO SRV domain set to confd.io
2013-11-03T19:04:53-08:00 confd[21356]: NOTICE Starting confd
2013-11-03T19:04:53-08:00 confd[21356]: NOTICE etcd nodes set to http://etcd0.confd.io:4001, http://etcd1.confd.io:4001
2013-11-03T19:04:54-08:00 confd[21356]: INFO /tmp/myconf2.conf has md5sum ae5c061f41de8895b6ef70803de9a455 should be 50d4ce679e1cf13e10cd9de90d258996
2013-11-03T19:04:54-08:00 confd[21356]: INFO Target config /tmp/myconf2.conf out of sync
2013-11-03T19:04:54-08:00 confd[21356]: INFO Target config /tmp/myconf2.conf has been updated 
```

By default the `etcd-scheme` will be set to http. If you would like to use https instead use the `-etcd-scheme` flag

```Bash
confd -onetime -verbose -srv-domain confd.io -etcd-scheme https
```

Both the SRV domain and etcd scheme can be configured in the confd configuration file. See the [Configuration Guide](https://github.com/kelseyhightower/confd/wiki/Configuration-Guide) for more details.
