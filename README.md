# confd

[![Build Status](https://travis-ci.org/kelseyhightower/confd.svg?branch=master)](https://travis-ci.org/kelseyhightower/confd)

`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date using data stored in [etcd](https://github.com/coreos/etcd),
  [consul](http://consul.io), [dynamodb](http://aws.amazon.com/dynamodb/), [redis](http://redis.io),
  [vault](https://vaultproject.io), [zookeeper](https://zookeeper.apache.org) or env vars and processing [template resources](docs/template-resources.md).
* reloading applications to pick up new config file changes

## this fork version
* Only some user-defined contents, such as template funcs

## Community

* IRC: `#confd` on Freenode
* Mailing list: [Google Groups](https://groups.google.com/forum/#!forum/confd-users)
* Website: [www.confd.io](http://www.confd.io)

## Building

Go 1.6 is required to build confd, which uses the new vendor directory.

```
$ mkdir -p $GOPATH/src/github.com/mafengwo
$ git clone https://github.com/mafengwo/confd.git $GOPATH/src/github.com/mafengwo/confd
$ cd $GOPATH/src/github.com/mafengwo/confd
$ ./build
```

You should now have confd in your `bin/` directory:

```
$ ls bin/
confd
```
## Starting

```
/usr/local/go/src/github.com/mafengwo/confd/bin/confd -backend etcd -interval 10 -node http://127.0.0.1:10001 -redisqueue 192.168.3.40:6379 &
```
## Getting Started

Before we begin be sure to [download and install confd](docs/installation.md).

* [quick start guide](docs/quick-start-guide.md)

## Next steps

Check out the [docs directory](docs) for more docs.
