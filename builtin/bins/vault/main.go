package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/vault"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		DatabaseFunc: vault.Database,
	})
}
