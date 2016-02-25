#!/bin/bash

cat <<EOT >> test.yaml
key: foobar
database:
  - host: 127.0.0.1
  - password: p@sSw0rd
  - port: "3306"
  - username: confd
upstream:
  - app1: 10.0.1.10:8080
  - app2: 10.0.1.11:8080
prefix:
  database:
    - host: 127.0.0.1
    - password: p@sSw0rd
    - port: "3306"
    - username: confd
  upstream:
    app1: 10.0.1.10:8080
    app2: 10.0.1.11:8080
EOT

# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --backend file --file test.yaml --watch
