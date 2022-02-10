#!/bin/bash

export DYNAMODB_VERSION=${1:-$DYNAMODB_VERSION}
export TMPDIR="/tmp/dynamodb"

sudo pip install awscli
command -v java 1>/dev/null || sudo apt install -y openjdk-17-jre

mkdir -p ${TMPDIR}

wget -v -O /tmp/dynamo.tar.gz https://s3.eu-central-1.amazonaws.com/dynamodb-local-frankfurt/dynamodb_local_${DYNAMODB_VERSION}.tar.gz
tar xzf /tmp/dynamo.tar.gz --directory ${TMPDIR}
java -Djava.library.path=${TMPDIR}/DynamoDBLocal_lib -jar ${TMPDIR}/DynamoDBLocal.jar -inMemory &