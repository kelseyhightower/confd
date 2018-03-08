#!/bin/bash

set -x
pushd $(dirname $0)

# Run the GCE metadata mocking server
go run main.go -port 8123 &

# Wait for it to come up
sleep 1
popd

# Needed env variables
export HOSTNAME=localhost

# Run confd with the google backend
which confd
confd \
    -onetime -watch \
    -log-level debug \
    -confdir ./integration/confdir \
    -node localhost:8123 \
    -backend google

# Clean up mocking servier
killall main
wait
