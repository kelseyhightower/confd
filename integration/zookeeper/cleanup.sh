#!/bin/bash

export ZOOKEEPER_VERSION=${1:-$ZOOKEEPER_VERSION}
export TMPDIR="/tmp/zookeeper"

${TMPDIR}/zookeeper-${ZOOKEEPER_VERSION}/bin/zkServer.sh stop
rm -rf ${TMPDIR}