# Logging

confd logs all error message to stderr and everything else to stdout. You can control the types of messages that get printed by using the `-debug`, `-verbose`, and `-quiet` flags and corresponding configuration file settings. See the [Configuration Guide](https://github.com/kelseyhightower/confd/wiki/Configuration-Guide) for more details.

Example log messages:

```Bash
2013-11-03T19:04:53-08:00 confd[21356]: INFO SRV domain set to confd.io
2013-11-03T19:04:53-08:00 confd[21356]: NOTICE Starting confd
2013-11-03T19:04:53-08:00 confd[21356]: NOTICE etcd nodes set to http://etcd0.confd.io:4001, http://etcd1.confd.io:4001
2013-11-03T19:04:54-08:00 confd[21356]: INFO /tmp/myconf2.conf has md5sum ae5c061f41de8895b6ef70803de9a455 should be 50d4ce679e1cf13e10cd9de90d258996
2013-11-03T19:04:54-08:00 confd[21356]: INFO Target config /tmp/myconf2.conf out of sync
2013-11-03T19:04:54-08:00 confd[21356]: INFO Target config /tmp/myconf2.conf has been updated
```
