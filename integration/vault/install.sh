#!/bin/bash

export VAULT_VERSION=${1:-$VAULT_VERSION}
export OS=$(go env GOOS)
export ARCH=$(go env GOARCH)

export TMPDIR="/tmp/vault"

mkdir -p ${TMPDIR}/bin
cd ${TMPDIR}

wget -q https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_${OS}_${ARCH}.zip
unzip -u -d ./bin vault_${VAULT_VERSION}_${OS}_${ARCH}.zip
./bin/vault server -dev &
