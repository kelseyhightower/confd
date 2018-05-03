#!/bin/bash

export HOSTNAME="localhost"
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

vault auth enable -path=test approle

echo 'path "*" {
  capabilities = ["read"]
}' > my-policy.hcl

vault write sys/policy/my-policy policy=@my-policy.hcl

vault write auth/test/role/my-role secret_id_ttl=120m token_num_uses=1000 token_ttl=60m token_max_ttl=120m secret_id_num_uses=10000 policies=my-policy

export ROLE_ID=$(vault read -field=role_id auth/test/role/my-role/role-id)
export SECRET_ID=$(vault write -f -field=secret_id auth/test/role/my-role/secret-id)

# Run confd
confd --onetime --log-level debug \
      --confdir ./integration/confdir \
      --backend vault \
      --auth-type app-role \
      --role-id $ROLE_ID \
      --secret-id $SECRET_ID \
      --vault-path=test \
      --node http://127.0.0.1:8200
