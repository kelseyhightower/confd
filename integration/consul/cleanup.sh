#!/bin/bash

export TMPDIR="/tmp/consul"

killall consul

rm -rf ${TMPDIR}