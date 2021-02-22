#!/bin/bash

export HOSTNAME="localhost"

vault secrets enable -version 1 -path kv-v1 kv

vault write kv-v1/exists key=foobar
vault write kv-v1/database host=127.0.0.1 port=3306 username=confd password=p@sSw0rd
vault write kv-v1/upstream app1=10.0.1.10:8080 app2=10.0.1.11:8080
vault write kv-v1/nested/east app1=10.0.1.10:8080
vault write kv-v1/nested/west app2=10.0.1.11:8080

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
      --confdir ./integration/vault-approle/confdir \
      --backend vault \
      --auth-type app-role \
      --role-id $ROLE_ID \
      --secret-id $SECRET_ID \
      --path=test \
      --node $VAULT_ADDR
