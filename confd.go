// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/confd/log"
	"path/filepath"
	"time"
)

func main() {
	if err := setConfig(); err != nil {
		log.Fatal(err.Error())
	}
	paths, err := confd.FindConfigs(config.Confd.ConfigDir)
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		for _, p := range paths {
			c, err := confd.NewConfig(p)
			if err != nil {
				log.Error(err.Error())
			}
			if err := c.Process(); err != nil {
				log.Error(err.Error())
			}
		}
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}
