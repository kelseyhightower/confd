# confd

[![Build Status](https://travis-ci.org/kelseyhightower/confd.svg?branch=master)](https://travis-ci.org/kelseyhightower/confd)

`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date using data stored in [etcd](https://github.com/coreos/etcd),
  [consul](http://consul.io), [dynamodb](http://aws.amazon.com/dynamodb/), [redis](http://redis.io),
  [vault](https://vaultproject.io), [zookeeper](https://zookeeper.apache.org) or env vars and processing [template resources](docs/template-resources.md).
* reloading applications to pick up new config file changes

## Community

* IRC: `#confd` on Freenode
* Mailing list: [Google Groups](https://groups.google.com/forum/#!forum/confd-users)
* Website: [www.confd.io](http://www.confd.io)

## Building

Go 1.6 is required to build confd, which uses the new vendor directory.

```
cd $GOPATH
go get github.com/kelseyhightower/confd
go build github.com/kelseyhightower/confd
```

You should now have confd binary in the working directory:

```
$ ls /
bin    confd    pkg    src
```

> *Note*: Using `go install github.com/kelseyhightower/confd` works too, and will place the binary into the ./bin/ folder.

### Building with the go tool

```
$ git clone https://github.com/kelseyhightower/confd.git
$ cd confd
```

```
$ pwd
/Users/kelseyhightower/confd
```

```
$ go build .
```

You can also produce a static binary too:

```
$ go build -a -installsuffix cgo -ldflags '-extld ld -extldflags -static' -a -x .
```

## Getting Started

Before we begin be sure to [download and install confd](docs/installation.md).

* [quick start guide](docs/quick-start-guide.md)

## Next steps

Check out the [docs directory](docs) for more docs.
