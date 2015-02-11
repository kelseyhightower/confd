#!/bin/bash

# feed zookeeper
export ZK_PATH="`dirname \"$0\"`"
sh -c "cd $ZK_PATH ; go run main.go"

# Run confd
./confd -verbose -debug -watch -confdir ./integration/confdir -backend zookeeper -node 127.0.0.1:2181
if [ $? -eq 1 ]
then
   echo good
else
   echo "watch blackhole spotted. Did you fix it ?"
fi
./confd -verbose -debug -confdir ./integration/confdir -interval 5 -backend zookeeper -node 127.0.0.1:2181

