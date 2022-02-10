#!/bin/bash

export TMPDIR="/tmp/dynamodb"

killall java
rm -rf ${TMPDIR}