package main

const Version = "0.12.0-dev"

// We want to replace this variable at build time with "-ldflags -X main.GitCommit=xxx", where const is not supported.
var GitCommit = ""
