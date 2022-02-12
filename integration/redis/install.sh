#!/bin/bash

export PORT=6379

set -e

sudo apt install -y redis-server
redis-server &

# Wait for server startup
timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost ${PORT}
