// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"github.com/BurntSushi/toml"
)

var (
	config     Config
	configFile = "/etc/confd/confd.toml"
)

type Config struct {
	Confd confdConfig
	Etcd  etcdConfig
}

type confdConfig struct {
	ConfigDir   string
	TemplateDir string
	Interval    int
}

type etcdConfig struct {
	Prefix   string
	Machines []string
}

func setConfig() error {
	etcdDefaults := etcdConfig{
		Prefix:   "/",
		Machines: []string{"http://127.0.0.1:4001"},
	}
	confdDefaults := confdConfig{
		Interval:    600,
		ConfigDir:   "/etc/confd/conf.d",
		TemplateDir: "/etc/confd/templates",
	}
	config.Etcd = etcdDefaults
	config.Confd = confdDefaults

	if IsFileExist(configFile) {
		_, err := toml.DecodeFile(configFile, &config)
		if err != nil {
			return err
		}
	}
	return nil
}
