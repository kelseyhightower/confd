#!/bin/bash

curl -L -X PUT http://127.0.0.1:4001/v2/keys/database/host -d value=127.0.0.1
curl -L -X PUT http://127.0.0.1:4001/v2/keys/database/password -d value=p@sSw0rd
curl -L -X PUT http://127.0.0.1:4001/v2/keys/database/port -d value=3306
curl -L -X PUT http://127.0.0.1:4001/v2/keys/database/username -d value=confd
curl -L -X PUT http://127.0.0.1:4001/v2/keys/upstream/app1 -d value=10.0.1.10:8080
curl -L -X PUT http://127.0.0.1:4001/v2/keys/upstream/app2 -d value=10.0.1.11:8080
curl -L -X PUT http://127.0.0.1:4001/v2/keys/prefix/database/host -d value=127.0.0.1
curl -L -X PUT http://127.0.0.1:4001/v2/keys/prefix/database/password -d value=p@sSw0rd
curl -L -X PUT http://127.0.0.1:4001/v2/keys/prefix/database/port -d value=3306
curl -L -X PUT http://127.0.0.1:4001/v2/keys/prefix/database/username -d value=confd
curl -L -X PUT http://127.0.0.1:4001/v2/keys/prefix/upstream/app1 -d value=10.0.1.10:8080
curl -L -X PUT http://127.0.0.1:4001/v2/keys/prefix/upstream/app2 -d value=10.0.1.11:8080


# Run confd
./confd -verbose -onetime -confdir ./integration/confdir -backend etcd -node 127.0.0.1:4001
