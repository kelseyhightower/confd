#!/bin/bash

export HOSTNAME="localhost"

cat <<EOT >> /tmp/confd-configMap.yaml
key: foobar
database:
  host: 127.0.0.1
  port: "3306"
  username: admin
upstream:
  app1: 10.0.1.10:8080
  app2: 10.0.1.11:8080
prefix:
  database:
    host: 127.0.0.1
    port: "3306"
    username: admin
  upstream:
    app1: 10.0.1.10:8080
    app2: 10.0.1.11:8080
nested:
  foo: bar
EOT

cat <<EOT >> /tmp/confd-secrets.yaml
database:
  password: p@sSw0rd
  username: confd
prefix:
  database:
    password: p@sSw0rd
    username: confd
nested:
  hip: hop
EOT

# Run confd
confd --onetime \
    --log-level debug \
    --confdir ./integration/confdir \
    --backend clconf \
    --file "/tmp/confd-configMap.yaml,/tmp/confd-secrets.yaml" \
    --watch
