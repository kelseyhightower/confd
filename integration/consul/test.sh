#!/bin/bash

# Configure consul
curl -X PUT http://127.0.0.1:8500/v1/kv/database/host -d '127.0.0.1'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/password -d 'p@sSw0rd'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/port -d '3306'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/username -d 'confd'
curl -X PUT http://127.0.0.1:8500/v1/kv/upstream/app1 -d '10.0.1.10:8080'
curl -X PUT http://127.0.0.1:8500/v1/kv/upstream/app2 -d '10.0.1.11:8080'
curl -X PUT http://127.0.0.1:8500/v1/kv/with_under_scores -d 'value_with_underscores'
curl -X PUT http://127.0.0.1:8500/v1/kv/path_here/with/under_scores -d 'value_path_with_underscores'

# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --backend consul --node 127.0.0.1:8500
