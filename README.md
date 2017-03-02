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
## Namespace
```
/*
 * 完成confd前的任务：根据etcd中的namespace将对应的标准配置文件和标准模版文件中的变量替换掉
 * 1.根据特殊的命名规则检查标准配置文件和模版文件是否存在
 * 2.用namespace替换变量生成一对临时的配置文件和模版文件，判断是否一样。如果不一样，替换成新的配置文件和模版文件
 *
 * 命令规范:
 * namespace - es-XXX-data, es-XXX-master, redis, memcached. (注：es-XXX-data和es-XXX-master只需生成一对配置文件和模版文件)(服务-业务-data || 服务-业务-master)
 * 标准配置文件 - es.tomlx, redis.tomlx(命名空间的替换词为 __NS__)
 * 标准模版文件 - es.tmplx, redis.tmplx(命名空间的替换词为 __NS__)
 * 配置文件 - es-XXX.toml
 * 模版文件 - es-XXX-tmpl
 */
```




## Getting Started

Before we begin be sure to [download and install confd](docs/installation.md).

* [quick start guide](docs/quick-start-guide.md)

## Next steps

Check out the [docs directory](docs) for more docs.
