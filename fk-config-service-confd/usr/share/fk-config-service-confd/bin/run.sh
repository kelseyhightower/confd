#!/bin/bash

set -e

PACKAGE="fk-config-service-confd"
USER="cfgsvc"
CONFD="/usr/share/$PACKAGE/bin/confd"

exec setuidgid $USER $CONFD 2>&1 >> /var/log/flipkart/config-service/confd-out.log

