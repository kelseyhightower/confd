#!/bin/bash

export HOSTNAME="localhost"

# Configure consul
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/key -d 'foobar'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/database/host -d '127.0.0.1'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/database/password -d 'p@sSw0rd'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/database/port -d '3306'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/database/username -d 'confd'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/upstream/app1 -d '10.0.1.10:8080'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/upstream/app2 -d '10.0.1.11:8080'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/nested/east/app1 -d '10.0.1.10:8080'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/nested/west/app2 -d '10.0.1.11:8080'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/database/host -d '127.0.0.1'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/database/password -d 'p@sSw0rd'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/database/port -d '3306'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/database/username -d 'confd'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/upstream/app1 -d '10.0.1.10:8080'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/upstream/app2 -d '10.0.1.11:8080'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/nested/east/app1 -d '10.0.1.10:8080'
curl -sX PUT http://$CONSUL_HOST:8500/v1/kv/prefix/nested/west/app2 -d '10.0.1.11:8080'

# Run confd
confd --onetime --log-level debug --confdir ./test/integration/confdir --backend consul --node $CONSUL_HOST:8500
