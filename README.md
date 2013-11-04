# confd

[![Build Status](https://travis-ci.org/kelseyhightower/confd.png?branch=master)](https://travis-ci.org/kelseyhightower/confd)

* IRC: `#confd` on Freenode
* Mailing list: [Google Groups](https://groups.google.com/forum/#!forum/confd-users)

`confd` is a lightweight configuration management tool focused on:

* keeping local configuration files up-to-date by polling [etcd](https://github.com/coreos/etcd) and processing [template resources](https://github.com/kelseyhightower/confd#template-resources).
* reloading applications to pick up new config file changes

## Getting Started

### Installing confd

Download the latest binary from [Github](https://github.com/kelseyhightower/confd/releases/tag/v0.1.2).

### Building

You can build confd from source:

```
git clone https://github.com/kelseyhightower/confd.git
cd confd
go build
```

This will produce the `confd` binary in the current directory.

## Usage

The following commands will process all the [template resources](https://github.com/kelseyhightower/confd#template-resources) found under `/etc/confd/conf.d`.

### Poll the etcd cluster in 30 second intervals

The "/production" string will be prefixed to keys when querying etcd at http://127.0.0.1:4001.

```
confd -i 30 -p '/production' -n 'http://127.0.0.1:4001'
```

### Single run without polling

Using default settings run one time and exit.

```
confd -onetime
```

### Client authentication

Same as above but authenticate with client certificates.

```
confd -onetime -key /etc/confd/ssl/client.key -cert /etc/confd/ssl/client.crt
```

## Configuration

See the [Configuration Guide](https://github.com/kelseyhightower/confd/wiki/Configuration-Guide)

## Template Resources

See [Template Resources](https://github.com/kelseyhightower/confd/wiki/Template-Resources)

## Templates

See [Templates](https://github.com/kelseyhightower/confd/wiki/Templates)

## Extras

### Configuration Management

- [confd-cookbook](https://github.com/rjocoleman/confd-cookbook)
