#!/bin/bash

export HOSTNAME="localhost"
mkdir backends1 backends2
cat <<EOT >> backends1/1.yaml
key: foobar
database:
  host: 127.0.0.1
  password: p@sSw0rd
  port: "3306"
  username: confd
EOT

cat <<EOT >> backends1/2.yaml
upstream:
  app1: 10.0.1.10:8080
  app2: 10.0.1.11:8080
EOT

cat <<EOT >> backends2/1.yaml
nested:
  app1: 10.0.1.10:8080
  app2: 10.0.1.11:8080
EOT

cat <<EOT >> backends2/2.yaml
prefix:
  database:
    host: 127.0.0.1
    password: p@sSw0rd
    port: "3306"
    username: confd
  upstream:
    app1: 10.0.1.10:8080
    app2: 10.0.1.11:8080
EOT

# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --backend file --file backends1/ --file backends2/ --watch

# Clean up after
rm -rf backends1 backends2
