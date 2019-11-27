#!/bin/bash

export HOSTNAME="localhost"

mkdir -p /database /upstream /nested/{east,west} /prefix/database /prefix/nested/{east,west} /prefix/upstream

echo "foobar" > /key
echo "127.0.0.1" > /database/host
echo "p@sSw0rd" > /database/password
echo "3306" > /database/port
echo "confd" > /database/username
echo "confd" > /database/username
echo "10.0.1.10:8080" > /upstream/app1
echo "10.0.1.11:8080" > /upstream/app2
echo "10.0.1.10:8080" > /nested/east/app1
echo "10.0.1.11:8080" > /nested/west/app2
echo "127.0.0.1" > /prefix/database/host
echo "p@sSw0rd" > /prefix/database/password
echo "3306" > /prefix/database/port
echo "confd" > /prefix/database/username
echo "10.0.1.10:8080" > /prefix/upstream/app1
echo "10.0.1.11:8080" > /prefix/upstream/app2
echo "10.0.1.10:8080" > /prefix/nested/east/app1
echo "10.0.1.11:8080" > /prefix/nested/west/app2

# Run confd
/mnt/c/tools/confd/bin/confd --onetime --log-level debug --confdir ./integration/confdir --backend filesystem --watch
