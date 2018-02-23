#!/bin/bash

export HOSTNAME="localhost"
export ETCDCTL_API="3"

etcdctl put /key foobar
etcdctl put /database/host 127.0.0.1
etcdctl put /database/password p@sSw0rd
etcdctl put /database/port 3306
etcdctl put /database/username confd
etcdctl put /upstream/app1 10.0.1.10:8080
etcdctl put /upstream/app2 10.0.1.11:8080
etcdctl put /prefix/database/host 127.0.0.1
etcdctl put /prefix/database/password p@sSw0rd
etcdctl put /prefix/database/port 3306
etcdctl put /prefix/database/username confd
etcdctl put /prefix/upstream/app1 10.0.1.10:8080
etcdctl put /prefix/upstream/app2 10.0.1.11:8080

# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --backend etcdv3 --node http://127.0.0.1:2379 --watch
