#!/bin/bash

redis-cli set /key foobar
redis-cli set /database/host 127.0.0.1
redis-cli set /database/password p@sSw0rd
redis-cli set /database/port 3306
redis-cli set /database/username confd
redis-cli set /upstream/app1 10.0.1.10:8080
redis-cli set /upstream/app2 10.0.1.11:8080
redis-cli set /prefix/database/host 127.0.0.1
redis-cli set /prefix/database/password p@sSw0rd
redis-cli set /prefix/database/port 3306
redis-cli set /prefix/database/username confd
redis-cli set /prefix/upstream/app1 10.0.1.10:8080
redis-cli set /prefix/upstream/app2 10.0.1.11:8080

# Run confd
./confd -verbose -debug -confdir ./integration/confdir -interval 5 -backend redis -node 127.0.0.1:6379
