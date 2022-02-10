#!/bin/bash

export TMPDIR="/tmp/rancher"

killall rancher-metadata

rm -rf ${TMPDIR}