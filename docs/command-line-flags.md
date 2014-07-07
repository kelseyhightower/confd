# Command Line Flags

Command line flags override the confd [configuration file](https://github.com/kelseyhightower/confd/wiki/Configuration-Guide).

```
confd -h
```

```Text
Usage of confd:
  -backend="": backend to use
  -backend-scheme="http": the backend URI scheme. (http or https)
  -client-ca-keys="": client ca keys
  -client-cert="": the client cert
  -client-key="": the client key
  -confdir="/etc/confd": confd conf directory
  -config-file="": the confd config file
  -debug=false: enable debug logging
  -interval=600: backend polling interval
  -node=[]: list of backend nodes
  -noop=false: only show pending changes, don't sync configs.
  -onetime=false: run once and exit
  -prefix="/": key path prefix
  -quiet=false: enable quiet logging. Only error messages are printed.
  -srv-domain="": the domain for the backend SRV record, i.e. example.com
  -verbose=false: enable verbose logging
  -version=false: print version and exit
```
