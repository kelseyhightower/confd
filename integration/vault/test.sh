#!/bin/bash

export HOSTNAME="localhost"
export VAULT_ADDR="http://127.0.0.1:8200/"
export ROOT_TOKEN="$(vault read -field id auth/token/lookup-self)"

vault secrets enable -path database kv
vault secrets enable -path key kv
vault secrets enable -path upstream kv
vault secrets enable -path nested kv

vault write key value=foobar
vault write database/host value=127.0.0.1
vault write database/port value=3306
vault write database/username value=confd
vault write database/password value=p@sSw0rd
vault write upstream app1=10.0.1.10:8080 app2=10.0.1.11:8080
vault write nested/east/app1 value=10.0.1.10:8080
vault write nested/west/app2 value=10.0.1.11:8080

# Run confd
confd --onetime --log-level debug \
      --confdir ./integration/confdir \
      --backend vault \
      --auth-type token \
      --auth-token $ROOT_TOKEN \
      --node http://127.0.0.1:8200
