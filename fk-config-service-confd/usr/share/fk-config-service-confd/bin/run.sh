#!/bin/bash

set -e

PACKAGE="fk-config-service-confd"
USER="cfgsvc"
CONFD="/usr/share/$PACKAGE/bin/confd"

exec $CONFD >> /var/log/flipkart/config-service/confd-out.log 2>&1

