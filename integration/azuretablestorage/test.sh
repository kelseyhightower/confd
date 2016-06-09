#!/bin/bash

export AZURE_STORAGE_ACCESS_KEY=YourAzureKeyGoesHere===
export AZURE_STORAGE_ACCOUNT=yourazurestorageaccount
export AZURE_STORAGE_TABLE=confd # Note: if this exists all data will be deleted

# feed azure table storage
# This creates the above table in the above account, and populates with test data
export AZ_PATH="`dirname \"$0\"`"
sh -c "cd $AZ_PATH ; go run main.go $AZURE_STORAGE_ACCOUNT $AZURE_STORAGE_ACCESS_KEY $AZURE_STORAGE_TABLE"

# Run confd, expect it to work
confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend azuretablestorage --table $AZURE_STORAGE_TABLE -storage-account $AZURE_STORAGE_ACCOUNT -client-key $AZURE_STORAGE_ACCESS_KEY
if [ $? -ne 0 ]
then
        exit 1
fi

# Run confd with --watch, expecting it to fail
confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend azuretablestorage --table $AZURE_STORAGE_TABLE -storage-account $AZURE_STORAGE_ACCOUNT -client-key $AZURE_STORAGE_ACCESS_KEY confd --watch
if [ $? -eq 0 ]
then
        exit 1
fi

# Run confd without Azure credentials, expecting it to fail
unset AZURE_STORAGE_ACCESS_KEY 
unset AZURE_STORAGE_ACCOUNT

confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend azuretablestorage --table $AZURE_STORAGE_TABLE -storage-account $AZURE_STORAGE_ACCOUNT -client-key $AZURE_STORAGE_ACCESS_KEY
if [ $? -eq 0 ]
then
        exit 1
fi
