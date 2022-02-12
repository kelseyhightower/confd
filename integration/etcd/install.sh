#!/bin/bash

set -e

export ETCD_VERSION=${1:-$ETCD_VERSION}
export ARCH=$(go env GOARCH)
export TMPDIR=/tmp/etcd
export PORT=2380

mkdir -p ${TMPDIR}/bin

cd ${TMPDIR}

wget -q https://storage.googleapis.com/etcd/v${ETCD_VERSION}/etcd-v${ETCD_VERSION}-$(go env GOOS)-${ARCH}.tar.gz
tar xzf etcd-v${ETCD_VERSION}-$(go env GOOS)-${ARCH}.tar.gz

mv etcd-v${ETCD_VERSION}-$(go env GOOS)-${ARCH}/etcd* ${TMPDIR}/bin
chmod 755 ${TMPDIR}/bin/etcd*

unset ETCD_VERSION

if [ $ARCH != "amd64" ]; then
    export ETCD_UNSUPPORTED_ARCH=${ARCH}
fi

${TMPDIR}/bin/etcd &

# Wait for server startup
timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost ${PORT}
