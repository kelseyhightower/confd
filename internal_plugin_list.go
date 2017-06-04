package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/env"
	"github.com/kelseyhightower/confd/plugin"
)

var InternalDatabases = map[string]plugin.DatabaseFunc{
	"env": env.Database,
}
