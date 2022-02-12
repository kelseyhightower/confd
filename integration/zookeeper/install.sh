#!/bin/bash

export ZOOKEEPER_VERSION=${1:-$ZOOKEEPER_VERSION}

export TMPDIR="/tmp/zookeeper"

cp integration/zookeeper/zoo.cfg /tmp/

mkdir -p ${TMPDIR}/bin
cd ${TMPDIR}

set -x
wget https://archive.apache.org/dist/zookeeper/zookeeper-${ZOOKEEPER_VERSION}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz -O ${TMPDIR}/zookeeper.tar.gz

#cp /tmp/zoo.cfg zookeeper-${ZOOKEEPER_VERSION}/conf/zoo.cfg

tar xzf ${TMPDIR}/zookeeper.tar.gz -C ${TMPDIR}

cp /tmp/zoo.cfg ${TMPDIR}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/conf
${TMPDIR}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/bin/zkServer.sh start
