package main

import (
	"github.com/kelseyhightower/confd/builtin/databases/redis"
	"github.com/kelseyhightower/confd/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Database: &redis.Client{},
	})
}
