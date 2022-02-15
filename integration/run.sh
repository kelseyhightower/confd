#!/bin/bash

set -e

# change curdir to reo root
cd $(dirname $0)/../

export DEBIAN_FRONTEND=noninteractive
export PATH=$(pwd)/bin:${PATH}

export CONSUL_VERSION="1.11.2"
export ETCD_VERSION="3.4.3"
export VAULT_VERSION="1.9.3"
export DYNAMODB_VERSION="latest"
export ZOOKEEPER_VERSION="3.6.3"
export RANCHER_VERSION="0.6.0"

#export INTEGRATION_TESTS=("file" "redis" "dynamodb" "zookeeper" "env" "consul" "etcd" ) # "vault")
export INTEGRATION_TESTS=("vault" "consul" "etcd" "redis" "env" "file" "dynamodb" "zookeeper")

apt -q update
apt install -y curl wget unzip python3-pip make git jq sudo psmisc netcat

make build

for t in ${INTEGRATION_TESTS[@]}; do
    echo "----------------------------------------"
    echo "<<< Running ${t} confd integration test ... >>>"

    if [ -x integration/${t}/install.sh ]; then
        echo "<<<<< Running Install: ${t}/install.sh script ... >>>>>";
        integration/${t}/install.sh || exit 1;
    fi

    for testfile in $(find integration/${t} -name test\*.sh); do
        if [ -x ${testfile} ]; then
            echo "<<<<< Running Test: ${testfile} script ... >>>>>";
            ${testfile} || exit 1;
            integration/expect/check.sh || exit 1;
            echo "<<<<< Tests finished ... >>>>>"
        fi
    done

    for cleanfile in $(find integration/${t} -name cleanup.sh); do
        if [ -x ${cleanfile} ]; then
            echo "<<<<< Running Cleanup: ${cleanfile} script ... >>>>>";
            ${cleanfile} || exit 1;
        fi
    done
    echo "----------------------------------------"
done


