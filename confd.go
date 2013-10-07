// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"flag"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
	"time"
)

func main() {
	log.Info("Starting confd")
	flag.Parse()
	if err := config.SetConfig(); err != nil {
		log.Fatal(err.Error())
	}
	for {
		if err := ProcessTemplateConfigs(); err != nil {
			log.Error(err.Error())
		}
		if config.Onetime() {
			break
		}
		time.Sleep(time.Duration(config.Interval()) * time.Second)
	}
}
