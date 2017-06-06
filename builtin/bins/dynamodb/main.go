package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/dynamodb"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		DatabaseFunc: dynamodb.Database,
	})
}
