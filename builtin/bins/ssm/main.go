package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/ssm"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Database: &ssm.Client{},
	})
}
