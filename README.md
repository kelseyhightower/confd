# confd

[![Build Status](https://travis-ci.org/kelseyhightower/confd.png?branch=master)](https://travis-ci.org/kelseyhightower/confd)

`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date by polling [etcd](https://github.com/coreos/etcd) or
  [Consul](http://consul.io) and processing [template resources](docs/template-resources.md).
* reloading applications to pick up new config file changes

## Community

* IRC: `#confd` on Freenode
* Mailing list: [Google Groups](https://groups.google.com/forum/#!forum/confd-users)
* Website: [www.confd.io](http://www.confd.io)

## Quick Start Guides

Before we begin be sure to [download and install confd](docs/installation.md).

* [etcd](docs/etcd-getting-started.md)
* [consul](docs/consul-getting-started.md)

## Next steps

Check out the [docs directory](docs) for more docs and [usage examples](docs/usage-examples.md).
