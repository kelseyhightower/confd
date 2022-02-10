#!/bin/bash

export RANCHER_VERSION=${1:-$RANCHER_VERSION}

export TMPDIR="/tmp/rancher"
cd ${TMPDIR}
mkdir -p ${TMPDIR}/rancher-metadata

wget -q https://github.com/rancher/rancher-metadata/releases/download/v${RANCHER_VERSION}/rancher-metadata.tar.gz
tar xzf rancher-metadata.tar.gz --strip-components=1 -C ${TMPDIR}/rancher-metadata
