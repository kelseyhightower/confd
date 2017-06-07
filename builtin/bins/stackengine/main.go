package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/stackengine"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		DatabaseFunc: stackengine.Database,
	})
}
