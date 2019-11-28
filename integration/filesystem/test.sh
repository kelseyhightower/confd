#!/bin/bash

export HOSTNAME="localhost"

sudo mkdir -p /database /upstream /nested/{east,west} /prefix/database /prefix/nested/{east,west} /prefix/upstream

echo "foobar" | sudo tee /key
echo "127.0.0.1" | sudo tee /database/host
echo "p@sSw0rd" | sudo tee /database/password
echo "3306" | sudo tee /database/port
echo "confd" | sudo tee /database/username
echo "confd" | sudo tee /database/username
echo "10.0.1.10:8080" | sudo tee /upstream/app1
echo "10.0.1.11:8080" | sudo tee /upstream/app2
echo "10.0.1.10:8080" | sudo tee /nested/east/app1
echo "10.0.1.11:8080" | sudo tee /nested/west/app2
echo "127.0.0.1" | sudo tee /prefix/database/host
echo "p@sSw0rd" | sudo tee /prefix/database/password
echo "3306" | sudo tee /prefix/database/port
echo "confd" | sudo tee /prefix/database/username
echo "10.0.1.10:8080" | sudo tee /prefix/upstream/app1
echo "10.0.1.11:8080" | sudo tee /prefix/upstream/app2
echo "10.0.1.10:8080" | sudo tee /prefix/nested/east/app1
echo "10.0.1.11:8080" | sudo tee /prefix/nested/west/app2

# Run confd
confd --onetime --log-level debug --confdir ./integration/confdir --backend filesystem --watch
