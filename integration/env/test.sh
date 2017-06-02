#!/bin/bash

export KEY="foobar"
export DATABASE_HOST="127.0.0.1"
export DATABASE_PASSWORD="p@sSw0rd"
export DATABASE_PORT="3306"
export DATABASE_USERNAME="confd"
export UPSTREAM_APP1="10.0.1.10:8080"
export UPSTREAM_APP2="10.0.1.11:8080"
export PREFIX_DATABASE_HOST="127.0.0.1"
export PREFIX_DATABASE_PASSWORD="p@sSw0rd"
export PREFIX_DATABASE_PORT="3306"
export PREFIX_DATABASE_USERNAME="confd"
export PREFIX_UPSTREAM_APP1="10.0.1.10:8080"
export PREFIX_UPSTREAM_APP2="10.0.1.11:8080"
export WITH_UNDER_SCORES="value_with_underscores"
export PATH_HERE_WITH_UNDER_SCORES="value_path_with_underscores"

confd --onetime --log-level debug --confdir ./integration/confdir --interval 5 --backend env
