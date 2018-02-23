#!/bin/bash

export HOSTNAME="localhost"

redis-cli set /key foobar
redis-cli set /database/host 127.0.0.1
redis-cli set /database/password p@sSw0rd
redis-cli set /database/port 3306
redis-cli set /database/username confd
redis-cli set /upstream/app1 10.0.1.10:8080
redis-cli set /upstream/app2 10.0.1.11:8080
redis-cli hset /prefix/database host 127.0.0.1
redis-cli hset /prefix/database password p@sSw0rd
redis-cli hset /prefix/database port 3306
redis-cli hset /prefix/database username confd
redis-cli hset /prefix/upstream app1 10.0.1.10:8080
redis-cli hset /prefix/upstream app2 10.0.1.11:8080

confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend redis --node 127.0.0.1:6379
if [ $? -ne 0 ]
then
        exit 1
fi

confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend redis --node 127.0.0.1:6379/0
if [ $? -ne 0 ]
then
        exit 1
fi
