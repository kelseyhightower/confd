#!/bin/bash

etcd_host=${1}
etcd_port=${2}

# Configure etcd
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/host -d value=127.0.0.1
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/password -d value=p@sSw0rd
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/port -d value=3306
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/username -d value=confd
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/upstream/app1 -d value=10.0.1.10:8080
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/upstream/app2 -d value=10.0.1.11:8080

# Run confd
./confd -watch -confdir ./integration/confdir -backend etcd -node "http://${etcd_host}:${etcd_port}" &
confd_pid=$!

sleep 1

curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/host -d value=127.9.9.9
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/password -d value=W4tch3Dp@sSw0rd
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/port -d value=3307
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/database/username -d value=watchedconfd
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/upstream/app1 -d value=20.0.2.20:8080
curl -L -X PUT http://${etcd_host}:${etcd_port}/v2/keys/upstream/app2 -d value=20.0.2.22:8080

sleep 1

kill ${confd_pid}
