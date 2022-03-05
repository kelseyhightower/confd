#!/bin/bash

export HOSTNAME="localhost"

redis-cli -h $REDIS_HOST set /key foobar
redis-cli -h $REDIS_HOST set /database/host 127.0.0.1
redis-cli -h $REDIS_HOST set /database/password p@sSw0rd
redis-cli -h $REDIS_HOST set /database/port 3306
redis-cli -h $REDIS_HOST set /database/username confd
redis-cli -h $REDIS_HOST set /upstream/app1 10.0.1.10:8080
redis-cli -h $REDIS_HOST set /upstream/app2 10.0.1.11:8080
redis-cli -h $REDIS_HOST hset /prefix/database host 127.0.0.1
redis-cli -h $REDIS_HOST hset /prefix/database password p@sSw0rd
redis-cli -h $REDIS_HOST hset /prefix/database port 3306
redis-cli -h $REDIS_HOST hset /prefix/database username confd
redis-cli -h $REDIS_HOST hset /prefix/upstream app1 10.0.1.10:8080
redis-cli -h $REDIS_HOST hset /prefix/upstream app2 10.0.1.11:8080

confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend redis --node $REDIS_HOST:6379
if [ $? -ne 0 ]
then
        exit 1
fi

confd --onetime --log-level debug --confdir ./test/integration/confdir --interval 5 --backend redis --node $REDIS_HOST:6379/0
if [ $? -ne 0 ]
then
        exit 1
fi
