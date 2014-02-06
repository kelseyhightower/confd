# Noop Mode

When in noop mode target configuration files will not be modified.

Noop mode can be enabled by passing the `-noop` flag on the command line:

```Bash
confd -noop
```

Or setting `noop` to true in the confd config file (i.e. /etc/confd/confd.toml). 

```TOML
noop = true
```

### Example usage

```Bash
confd -onetime -noop -verbose
```

Output:
```Bash
2013-11-03T18:55:38-08:00 confd[21353]: NOTICE Starting confd
2013-11-03T18:55:38-08:00 confd[21353]: NOTICE etcd nodes set to http://127.0.0.1:4001
2013-11-03T18:55:38-08:00 confd[21353]: INFO /tmp/myconf2.conf has md5sum ae5c061f41de8895b6ef70803de9a455 should be 50d4ce679e1cf13e10cd9de90d258996
2013-11-03T18:55:38-08:00 confd[21353]: WARNING Noop mode enabled /tmp/myconf2.conf will not be modified
```
