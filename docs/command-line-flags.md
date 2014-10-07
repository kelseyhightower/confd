# Command Line Flags

Command line flags override the confd [configuration file](configuration-guide.md).

```
confd -h
```

```Text
Usage of confd:
  -backend="etcd": backend to use
  -client-ca-keys="": client ca keys
  -client-cert="": the client cert
  -client-key="": the client key
  -confdir="/etc/confd": confd conf directory
  -config-file="": the confd config file
  -debug=false: enable debug logging
  -interval=600: backend polling interval
  -keep-stage-file=false: keep staged files
  -node=[]: list of backend nodes
  -noop=false: only show pending changes
  -onetime=false: run once and exit
  -prefix="/": key path prefix
  -quiet=false: enable quiet logging
  -scheme="http": the backend URI scheme (http or https)
  -srv-domain="": the name of the resource record
  -verbose=false: enable verbose logging
  -version=false: print version and exit
```

> The -scheme flag is only used to set the URL scheme for nodes retrieved from DNS SRV records.
