package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/etcd"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		DatabaseFunc: etcd.Database,
	})
}
