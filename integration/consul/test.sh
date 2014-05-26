#!/bin/bash

docker_host=${1}
consul_id=$(docker run -p 8500:8500 -p 8400:8400 -d kelseyhightower/consul -server -bootstrap -ui-dir /opt/consul/web_ui -client 0.0.0.0 -advertise ${docker_host})

# Configure consul
curl -X PUT http://${docker_host}:8500/v1/kv/database/host -d '127.0.0.1'
curl -X PUT http://${docker_host}:8500/v1/kv/database/password -d 'p@sSw0rd'
curl -X PUT http://${docker_host}:8500/v1/kv/database/port -d '3306'
curl -X PUT http://${docker_host}:8500/v1/kv/database/username -d 'confd'
curl -X PUT http://${docker_host}:8500/v1/kv/upstream/app1 -d '10.0.1.10:8080'
curl -X PUT http://${docker_host}:8500/v1/kv/upstream/app2 -d '10.0.1.11:8080'

# Run confd
./confd -onetime -confdir ./integration/confdir -backend consul -consul-addr "${1}:8500"

# Cleanup
docker stop $consul_id
