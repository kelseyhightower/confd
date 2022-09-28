# confd

[![Build Status](https://travis-ci.org/kelseyhightower/confd.svg?branch=master)](https://travis-ci.org/kelseyhightower/confd)

`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date using data stored in [etcd](https://github.com/coreos/etcd),
  [consul](http://consul.io), [dynamodb](http://aws.amazon.com/dynamodb/), [redis](http://redis.io),
  [vault](https://vaultproject.io), [zookeeper](https://zookeeper.apache.org), [aws ssm parameter store](https://aws.amazon.com/ec2/systems-manager/) or env vars and processing [template resources](docs/template-resources.md).
* reloading applications to pick up new config file changes

## Project Status

`confd` is currently being cleaned up to build on later versions of Go and moving to adopt native support for [Go modules](https://go.dev/blog/using-go-modules). As part of this work the following major changes are being made:

* The `etcd` and `etcdv3` backend are going to be merged. etcd v2 has been deprecated and both backend will now use etcdv3 client libraries.
* The `cget`, `cgets`, `cgetv`, and `cgetvs` templates function have been removed due to an unmaintained dependency `github.com/xordataexchange/crypt/encoding/secconf`. We need to rethink encryption in the core project and rely only on the standard library going forward. In the meanwhile these template function will not work and if support is required you will need to stick with an older version of confd.

## Community

* IRC: `#confd` on Freenode
* Mailing list: [Google Groups](https://groups.google.com/forum/#!forum/confd-users)
* Website: [www.confd.io](http://www.confd.io)

## Building

Go 1.10 is required to build confd, which uses the new vendor directory.

```
$ mkdir -p $GOPATH/src/github.com/kelseyhightower
$ git clone https://github.com/kelseyhightower/confd.git $GOPATH/src/github.com/kelseyhightower/confd
$ cd $GOPATH/src/github.com/kelseyhightower/confd
$ make
```

You should now have confd in your `bin/` directory:

```
$ ls bin/
confd
```

## Getting Started

Before we begin be sure to [download and install confd](docs/installation.md).

* [quick start guide](docs/quick-start-guide.md)

## Next steps

Check out the [docs directory](docs) for more docs.
