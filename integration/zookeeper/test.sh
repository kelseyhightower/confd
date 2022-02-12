#!/bin/bash

export HOSTNAME="localhost"
export PORT=2181

while ! nc -z localhost ${PORT}; do
  sleep 1 # wait for 1 second before check again
done

set -e


# feed zookeeper
export ZK_PATH="`dirname \"$0\"`"
sh -c "cd $ZK_PATH; go run main.go"

# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend zookeeper --node 127.0.0.1:2181 -watch
