package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/etcdv3"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Database: &etcdv3.Client{},
	})
}
