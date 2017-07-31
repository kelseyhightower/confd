#!/bin/sh
go run confd.go config.go version.go node_var.go \
    -backend kubernetes \
    -node 127.0.0.1:4000 \
    -scheme http \
    -auth-token TOKEN12345 \
    -interval 5 \
    -log-level info
