// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"github.com/kelseyhightower/confd/log"
	"path/filepath"
	"time"
)

func main() {
	if err := setConfig(); err != nil {
		log.Fatal(err.Error())
	}
	paths, err := filepath.Glob(filepath.Join(config.Confd.ConfigDir, "*json"))
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		for _, p := range paths {
			ct, err := NewConfigTemplateFromFile(p)
			if err != nil {
				log.Error(err.Error())
			}
			if err := ct.Process(); err != nil {
				log.Error(err.Error())
			}
		}
		interval := config.Confd.Interval
		if err != nil {
			log.Fatal(err.Error())
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
