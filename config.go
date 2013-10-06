package main

import (
	"github.com/BurntSushi/toml"
)

var (
	config     Config
	configFile = "/etc/confd/confd.toml"
)

type Config struct {
	Confd confd
	Etcd  etcd
}

type confd struct {
	ConfigDir   string
	TemplateDir string
	Interval    int
}

type etcd struct {
	Prefix   string
	Machines []string
}

func setConfig() error {
	config.Etcd = etcd{
		Prefix:   "/",
		Machines: []string{"http://127.0.0.1:4001"},
	}
	config.Confd = confd{
		Interval:    600,
		ConfigDir:   "/etc/confd/conf.d",
		TemplateDir: "/etc/confd/templates",
	}
	if IsFileExist(configFile) {
		_, err := toml.DecodeFile(configFile, &config)
		if err != nil {
			return err
		}
	}
	return nil
}
