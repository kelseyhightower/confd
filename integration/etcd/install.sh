#!/bin/bash

export ETCD_VERSION=${1:-$ETCD_VERSION}
export ARCH=$(go env GOARCH)
export TMPDIR=/tmp/etcd

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