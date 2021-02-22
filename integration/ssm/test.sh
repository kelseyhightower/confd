#!/bin/bash -x

export HOSTNAME="localhost"
export AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID"
export AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY"
export AWS_DEFAULT_REGION="$AWS_DEFAULT_REGION"
export AWS_REGION="$AWS_REGION"

aws ssm put-parameter --name "/key" --type "String" --value "foobar" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/database/host" --type "String" --value "127.0.0.1" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/database/password" --type "String" --value "p@sSw0rd" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/database/port" --type "String" --value "3306" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/database/username" --type "String" --value "confd" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/upstream/app1" --type "String" --value "10.0.1.10:8080" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/upstream/app2" --type "String" --value "10.0.1.11:8080" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/prefix/database/host" --type "String" --value "127.0.0.1" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/prefix/database/password" --type "String" --value "p@sSw0rd" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/prefix/database/port" --type "String" --value "3306" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/prefix/database/username" --type "String" --value "confd" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/prefix/upstream/app1" --type "String" --value "10.0.1.10:8080" --endpoint-url $SSM_ENDPOINT_URL
aws ssm put-parameter --name "/prefix/upstream/app2" --type "String" --value "10.0.1.11:8080" --endpoint-url $SSM_ENDPOINT_URL

# Run confd, expect it to work
confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend ssm
if [ $? -ne 0 ]
then
    exit 1
fi

# Run confd with --watch, expecting it to fail
confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend ssm --watch
if [ $? -eq 0 ]
then
    exit 1
fi

# Run confd without AWS credentials, expecting it to fail
unset AWS_ACCESS_KEY_ID
unset AWS_SECRET_ACCESS_KEY

confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend ssm
if [ $? -eq 0 ]
then
    exit 1
fi