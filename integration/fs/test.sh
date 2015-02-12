#!/bin/bash

# get full path: http://stackoverflow.com/questions/4774054/reliable-way-for-a-bash-script-to-get-the-full-path-to-itself
pushd `dirname $0` > /dev/null
ROOTPATH=`pwd -P`
popd > /dev/null

mkdir -p $ROOTPATH/database
mkdir -p $ROOTPATH/upstream
mkdir -p $ROOTPATH/prefix/database
mkdir -p $ROOTPATH/prefix/upstream

printf "foobar"         > $ROOTPATH/key
printf "127.0.0.1"      > $ROOTPATH/database/host
printf "p@sSw0rd"       > $ROOTPATH/database/password
printf "3306"           > $ROOTPATH/database/port
printf "confd"          > $ROOTPATH/database/username
printf "10.0.1.10:8080" > $ROOTPATH/upstream/app1
printf "10.0.1.11:8080" > $ROOTPATH/upstream/app2
printf "127.0.0.1"      > $ROOTPATH/prefix/database/host
printf "p@sSw0rd"       > $ROOTPATH/prefix/database/password
printf "3306"           > $ROOTPATH/prefix/database/port
printf "confd"          > $ROOTPATH/prefix/database/username
printf "10.0.1.10:8080" > $ROOTPATH/prefix/upstream/app1
printf "10.0.1.11:8080" > $ROOTPATH/prefix/upstream/app2

# Run confd with --watch, expecting it to fail
confd --onetime --log-level debug --confdir ./integration/confdir --backend fs --fs-rootpath $ROOTPATH --watch
if [ $? -eq 0 ]
then
	exit 1
fi
confd --onetime --log-level debug --confdir ./integration/confdir --backend fs --fs-rootpath $ROOTPATH
