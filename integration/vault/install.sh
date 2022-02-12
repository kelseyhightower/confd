#!/bin/bash

set -e

export VAULT_VERSION=${1:-$VAULT_VERSION}
export OS=$(go env GOOS)
export ARCH=$(go env GOARCH)

export TMPDIR="/tmp/vault"
export PORT=8200
export VAULT_ADDR="http://127.0.0.1:8200/"

mkdir -p ${TMPDIR}/bin
cd ${TMPDIR}

wget -q https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_${OS}_${ARCH}.zip
unzip -u -d ./bin vault_${VAULT_VERSION}_${OS}_${ARCH}.zip
./bin/vault server -dev &

# Wait for server startup
timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost ${PORT}

./bin/vault secrets enable -path database kv
./bin/vault secrets enable -path key kv
./bin/vault secrets enable -path upstream kv
./bin/vault secrets enable -path nested kv
