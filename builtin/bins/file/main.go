package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/file"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Database: &file.Client{},
	})
}
