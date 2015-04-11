package main

import (
	"os"
	"reflect"

	"github.com/codegangsta/cli"
	"github.com/kelseyhightower/confd"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/util"
)

var (
	globalConfig    *config.GlobalConfig     = nil
	templateConfigs []*config.TemplateConfig = nil
)

func handleGlobalAndTemplateOpts(globalFlags []Flag, templateFlags []Flag, c *cli.Context) error {
	var err error = nil

	// global configuration
	var gc *config.GlobalConfig = config.NewGlobalConfig()
	if c.GlobalIsSet("config-file") {
		gc, err = util.GetGlobalConfig(c.GlobalString("config-file"))
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	overwriteWithCliFlags(globalFlags, c, true, gc)

	// if conf directory was provided explicitly, use it
	var tcs []*config.TemplateConfig = make([]*config.TemplateConfig, 0)
	if gc.ConfDir != "" {
		tcs, err = util.GetTemplateConfigs(gc.ConfDir)
		if err != nil {
			log.Fatal(err.Error())
		}
		// if no template configs were found, nothing to do.
		if len(tcs) == 0 {
			log.Warning("no template configurations were found in: " + gc.ConfDir)
			os.Exit(0)
		}
	} else {
		var tc *config.TemplateConfig = config.NewTemplateConfig()
		overwriteWithCliFlags(templateFlags, c, true, tc)
		tcs = append(tcs, tc)
	}

	globalConfig = gc
	templateConfigs = tcs

	return nil
}

func handleBackendOptsAndRun(backendFlags []Flag, c *cli.Context) {
	// set requested command/backend
	backend := c.Command.Name

	// backend configuration
	backendConfig := config.NewBackendConfig(backend)
	if c.GlobalIsSet("config-file") {
		bc, err := util.GetBackendConfig(backend, c.GlobalString("config-file"))
		if err == nil {
			backendConfig = bc
		}
	}
	overwriteWithCliFlags(backendFlags, c, false, backendConfig)

	// run!
	confd.Run(globalConfig, templateConfigs, backendConfig)
}

func main() {
	// lookup flags
	globalFlags := getFlagsFromType(reflect.TypeOf((*config.GlobalConfig)(nil)).Elem())
	templateFlags := getFlagsFromType(reflect.TypeOf((*config.TemplateConfig)(nil)).Elem())
	consulFlags := getFlagsFromType(reflect.TypeOf((*config.ConsulBackendConfig)(nil)).Elem())
	envFlags := getFlagsFromType(reflect.TypeOf((*config.EnvBackendConfig)(nil)).Elem())
	etcdFlags := getFlagsFromType(reflect.TypeOf((*config.EtcdBackendConfig)(nil)).Elem())
	redisFlags := getFlagsFromType(reflect.TypeOf((*config.RedisBackendConfig)(nil)).Elem())
	zookeeperFlags := getFlagsFromType(reflect.TypeOf((*config.ZookeeperBackendConfig)(nil)).Elem())
	dynamodbFlags := getFlagsFromType(reflect.TypeOf((*config.DynamoDBBackendConfig)(nil)).Elem())
	fsFlags := getFlagsFromType(reflect.TypeOf((*config.FsBackendConfig)(nil)).Elem())

	// app
	app := cli.NewApp()
	app.Name = "confd"
	app.Version = "0.11.0-dev"
	app.Usage = "manage local application configuration files using templates and data"
	app.Flags = append(
		[]cli.Flag{
			cli.StringFlag{
				Name:  "config-file",
				Value: "/etc/confd/confd.toml",
				Usage: "the confd config file",
			},
		},
		append(getCliFlags(globalFlags), getCliFlags(templateFlags)...)...,
	)
	app.Before = func(c *cli.Context) error {
		return handleGlobalAndTemplateOpts(globalFlags, templateFlags, c)
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "consul",
			Flags:  getCliFlags(consulFlags),
			Action: func(c *cli.Context) { handleBackendOptsAndRun(consulFlags, c) },
		},
		cli.Command{
			Name:   "env",
			Flags:  getCliFlags(envFlags),
			Action: func(c *cli.Context) { handleBackendOptsAndRun(envFlags, c) },
		},
		cli.Command{
			Name:   "etcd",
			Flags:  getCliFlags(etcdFlags),
			Action: func(c *cli.Context) { handleBackendOptsAndRun(etcdFlags, c) },
		},
		cli.Command{
			Name:   "redis",
			Flags:  getCliFlags(redisFlags),
			Action: func(c *cli.Context) { handleBackendOptsAndRun(redisFlags, c) },
		},
		cli.Command{
			Name:   "zookeeper",
			Flags:  getCliFlags(zookeeperFlags),
			Action: func(c *cli.Context) { handleBackendOptsAndRun(zookeeperFlags, c) },
		},
		cli.Command{
			Name:   "dynamodb",
			Flags:  getCliFlags(dynamodbFlags),
			Action: func(c *cli.Context) { handleBackendOptsAndRun(dynamodbFlags, c) },
		},
		cli.Command{
			Name:   "fs",
			Flags:  getCliFlags(fsFlags),
			Action: func(c *cli.Context) { handleBackendOptsAndRun(fsFlags, c) },
		},
	}
	app.Run(os.Args)
}
