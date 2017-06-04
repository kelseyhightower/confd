package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/env"
	"github.com/kelseyhightower/confd/confd"
)

var InternalDatabases = map[string]confd.Database{
	"env": &env.Client{},
}
