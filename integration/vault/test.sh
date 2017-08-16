#!/bin/bash
set -e
ROOT_TOKEN=$(vault read -field id auth/token/lookup-self)

vault mount -path database generic
vault mount -path key generic
vault mount -path upstream generic
vault mount -path with_under_scores generic
vault mount -path path_here generic

vault write key value=foobar
vault write database/host value=127.0.0.1
vault write database/port value=3306
vault write database/username value=confd
vault write database/password value=p@sSw0rd
vault write upstream app1=10.0.1.10:8080 app2=10.0.1.11:8080
vault write with_under_scores value=value_with_underscores
vault write path_here/with/under_scores value=value_path_with_underscores

# Run confd
confd --onetime --log-level debug \
      --confdir ./integration/confdir \
      --backend vault \
      --auth-type token \
      --auth-token $ROOT_TOKEN \
      --node http://127.0.0.1:8200
