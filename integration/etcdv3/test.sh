#!/bin/bash

ETCDCTL_API=3 etcdctl put /key foobar
ETCDCTL_API=3 etcdctl put /database/host 127.0.0.1
ETCDCTL_API=3 etcdctl put /database/password p@sSw0rd
ETCDCTL_API=3 etcdctl put /database/port 3306
ETCDCTL_API=3 etcdctl put /database/username confd
ETCDCTL_API=3 etcdctl put /upstream/app1 10.0.1.10:8080
ETCDCTL_API=3 etcdctl put /upstream/app2 10.0.1.11:8080
ETCDCTL_API=3 etcdctl put /prefix/database/host 127.0.0.1
ETCDCTL_API=3 etcdctl put /prefix/database/password p@sSw0rd
ETCDCTL_API=3 etcdctl put /prefix/database/port 3306
ETCDCTL_API=3 etcdctl put /prefix/database/username confd
ETCDCTL_API=3 etcdctl put /prefix/upstream/app1 10.0.1.10:8080
ETCDCTL_API=3 etcdctl put /prefix/upstream/app2 10.0.1.11:8080


# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --backend etcdv3 --node http://127.0.0.1:2379 --watch
