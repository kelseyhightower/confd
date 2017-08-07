package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/rancher"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Database: &rancher.Client{},
	})
}
