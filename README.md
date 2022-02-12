**_Note: This is a divergent fork of [kelseyhightower/confd](https://github.com/kelseyhightower/confd). Backward compatibility is not guaranteed. YMMV_**

# confd

[![Integration Tests](https://github.com/abtreece/confd/actions/workflows/integration-tests.yml/badge.svg)](https://github.com/abtreece/confd/actions/workflows/integration-tests.yml)
[![CodeQL](https://github.com/abtreece/confd/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/abtreece/confd/actions/workflows/codeql-analysis.yml)


`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date using data stored in [etcd](https://github.com/etcd-io/etcd),
  [consul](http://consul.io), [dynamodb](http://aws.amazon.com/dynamodb/), [redis](http://redis.io),
  [vault](https://vaultproject.io), [zookeeper](https://zookeeper.apache.org), [aws ssm parameter store](https://aws.amazon.com/ec2/systems-manager/) or env vars and processing [template resources](docs/template-resources.md).
* reloading applications to pick up new config file changes

## Community


## Building

Go 1.12 is required to build confd, which uses Go Modules

```
$ git clone https://github.com/abtreece/confd.git
$ cd confd
$ make
```

You should now have `confd` in your `bin/` directory:

```
$ ls bin/
confd
```

## Getting Started

Before we begin be sure to [download and install confd](docs/installation.md).

* [quick start guide](docs/quick-start-guide.md)

## Next steps

Check out the [docs directory](docs) for more docs.
