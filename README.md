# confd

[![Build Status](https://travis-ci.org/bacongobbler/confd.svg?branch=master)](https://travis-ci.org/bacongobbler/confd)

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
$ mkdir -p $GOPATH/src/github.com/bacongobbler
$ git clone https://github.com/bacongobbler/confd.git $GOPATH/src/github.com/bacongobbler/confd
$ cd $GOPATH/src/github.com/bacongobbler/confd
$ ./build
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
