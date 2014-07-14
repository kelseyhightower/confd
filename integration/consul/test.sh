#!/bin/bash

# Configure consul
curl -X PUT http://127.0.0.1:8500/v1/kv/database/host -d '127.0.0.1'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/password -d 'p@sSw0rd'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/port -d '3306'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/username -d 'confd'
curl -X PUT http://127.0.0.1:8500/v1/kv/upstream/app1 -d '10.0.1.10:8080'
curl -X PUT http://127.0.0.1:8500/v1/kv/upstream/app2 -d '10.0.1.11:8080'

curl -X PUT http://127.0.0.1:8500/v1/agent/service/register -d '{"ID":"rails1", "Name":"rails","Tags":["master", "v1"],"Port":8080,"Check":{}}'
curl -X PUT http://127.0.0.1:8500/v1/agent/service/register -d '{"ID":"rails2", "Name":"rails","Tags":["master", "v2"],"Port":9090,"Check":{}}'
curl -X PUT http://127.0.0.1:8500/v1/agent/service/register -d '{"ID":"rails3", "Name":"rails","Tags":["slave", "v2"],"Port":9999,"Check":{"Script": "curl localhost:9999", "Interval": "10s", "TTL": "15s"}}'

# Run confd
./confd -onetime -verbose -confdir ./integration/confdir -backend consul -node "127.0.0.1:8500"
