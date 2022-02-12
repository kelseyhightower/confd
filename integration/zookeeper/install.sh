#!/bin/bash

export ZOOKEEPER_VERSION=${1:-$ZOOKEEPER_VERSION}
export TMPDIR="/tmp/zookeeper"
export PORT=2181

cp integration/zookeeper/zoo.cfg /tmp/

mkdir -p ${TMPDIR}/bin
cd ${TMPDIR}

command -v java 1>/dev/null || sudo apt install -y openjdk-17-jre

wget https://archive.apache.org/dist/zookeeper/zookeeper-${ZOOKEEPER_VERSION}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz -O ${TMPDIR}/zookeeper.tar.gz

#cp /tmp/zoo.cfg zookeeper-${ZOOKEEPER_VERSION}/conf/zoo.cfg

tar xzf ${TMPDIR}/zookeeper.tar.gz -C ${TMPDIR}

cp /tmp/zoo.cfg ${TMPDIR}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/conf
${TMPDIR}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/bin/zkServer.sh start

if [ $? -eq 1 ]; then
    cat ${TMPDIR}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin/logs/*.out;
    ps ax | grep zoo

    cat /tmp/zookeeper*/*.pid
fi

# Wait for server startup
timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost ${PORT}
