#!/bin/bash

set -e

export CONSUL_VERSION=${1:-$CONSUL_VERSION}
export OS=$(go env GOOS)
export ARCH=$(go env GOARCH)
export TMPDIR="/tmp/consul"
export PORT=8500

mkdir -p ${TMPDIR}/bin
cd ${TMPDIR}

wget -q https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_${OS}_${ARCH}.zip
unzip -d ./bin consul_${CONSUL_VERSION}_${OS}_${ARCH}.zip
./bin/consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul -bind 127.0.0.1 &

# Wait for server startup
timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost ${PORT}
