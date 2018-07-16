#!/bin/bash

export HOSTNAME="localhost"
export SECRETSMANAGER_LOCAL="1"
export AWS_ACCESS_KEY_ID="foo"
export AWS_SECRET_ACCESS_KEY="bar"
export AWS_DEFAULT_REGION="us-east-1"
export AWS_REGION="us-east-1"
export SECRETSMANAGER_ENDPOINT_URL="http://localhost:8002"

# aws secretsmanager describe-secret --secret-id tutorials/MyFirstTutorialSecret
aws secretsmanager  create-secret --name "/key" --secret-string "foobar" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/database/host" --secret-string "127.0.0.1" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/database/password" --secret-string "p@sSw0rd" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/database/port" --secret-string "3306" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/database/username" --secret-string "confd" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/upstream/app1" --secret-string "10.0.1.10:8080" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/upstream/app2" --secret-string "10.0.1.11:8080" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/prefix/database/host" --secret-string "127.0.0.1" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/prefix/database/password" --secret-string "p@sSw0rd" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/prefix/database/port" --secret-string "3306" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/prefix/database/username" --secret-string "confd" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/prefix/upstream/app1" --secret-string "10.0.1.10:8080" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL
aws secretsmanager  create-secret --name "/prefix/upstream/app2" --secret-string "10.0.1.11:8080" --endpoint-url $SECRETSMANAGER_ENDPOINT_URL

# Run confd, expect it to work
confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend secretsmanager --table confd
if [ $? -ne 0 ]
then
        exit 1
fi

# Run confd with --watch, expecting it to fail
confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend secretsmanager --table confd --watch
if [ $? -eq 0 ]
then
        exit 1
fi
