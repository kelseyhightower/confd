package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/zookeeper"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Database: &zookeeper.Client{},
	})
}
