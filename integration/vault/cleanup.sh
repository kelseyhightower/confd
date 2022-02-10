#!/bin/bash

export TMPDIR="/tmp/vault"

killall vault
rm -rf ${TMPDIR}