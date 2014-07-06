# Command Line Flags

Command line flags override the confd [configuration file](https://github.com/kelseyhightower/confd/wiki/Configuration-Guide), which is useful for testing and onetime runs. This also means that you can use confd without a configuration file.

> Note: All flags use a single `-`, which is much different from GNU style flags that use `--`. So you have to use `-noop` and not `--noop`.

To display the confd command line flags use the `-h` flag

```Bash
confd -h
Usage of confd:
  -client-cert="": the client cert
  -client-key="": the client key
  -confdir="/etc/confd": confd conf directory
  -config-file="": the confd config file
  -backend="": configuration backend to use. (consul, etcd, or env)
  -debug=false: enable debug logging
  -interval=600: etcd polling interval
  -node=[]: list of etcd nodes
  -noop=false: only show pending changes, don't sync configs.
  -onetime=false: run once and exit
  -prefix="/": etcd key path prefix
  -quiet=false: enable quiet logging. Only error messages are printed.
  -srv-domain="": the domain to query for the backend SRV record, i.e. example.com
  -verbose=false: enable verbose logging
```
