#!/bin/bash

export PATH=/tmp/vault/bin:$PATH
export HOSTNAME="localhost"

export VAULT_ADDR="http://127.0.0.1:8200/"
export ROOT_TOKEN="$(vault read -field id auth/token/lookup-self)"

set -e

vault secrets enable -path=pki pki
vault secrets tune -max-lease-ttl=10h pki

vault write key/key value=foobar
vault write database/host value=127.0.0.1
vault write database/port value=3306
vault write database/username value=confd
vault write database/password value=p@sSw0rd
vault write upstream/app1 value=10.0.1.10:8080
vault write upstream/app2 value=10.0.1.11:8080
vault write nested/east/app1 value=10.0.1.10:8080
vault write nested/west/app2 value=10.0.1.11:8080

vault write pki/root/generate/internal \
    common_name=example.com \
    ttl=8760h
vault write pki/config/urls \
    issuing_certificates="${VAULT_ADDR}/v1/pki/ca" \
    crl_distribution_points="${VAULT_ADDR}/v1/pki/crl"
vault write pki/roles/my-role \
    allowed_domains=example.com \
    allow_subdomains=true \
    max_ttl=8h

# Run confd
confd --onetime --log-level debug \
      --confdir ./integration/confdir \
      --backend vaultpki \
      --auth-type token \
      --auth-token $ROOT_TOKEN \
      --node http://127.0.0.1:8200

vault delete pki/root
vault secrets disable pki
