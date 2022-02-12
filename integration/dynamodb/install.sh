#!/bin/bash

set -e

export DYNAMODB_VERSION=${1:-$DYNAMODB_VERSION}
export TMPDIR="/tmp/dynamodb"
export PORT=8000

sudo pip install awscli
command -v java 1>/dev/null || sudo apt install -y openjdk-17-jre

mkdir -p ${TMPDIR}

wget -v -O /tmp/dynamo.tar.gz https://s3.eu-central-1.amazonaws.com/dynamodb-local-frankfurt/dynamodb_local_${DYNAMODB_VERSION}.tar.gz
tar xzf /tmp/dynamo.tar.gz --directory ${TMPDIR}
java -Djava.library.path=${TMPDIR}/DynamoDBLocal_lib -jar ${TMPDIR}/DynamoDBLocal.jar -inMemory &

# Wait for server startup
timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost ${PORT}
