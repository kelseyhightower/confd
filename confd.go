// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"flag"
	"github.com/kelseyhightower/confd/log"
	"time"
)

func main() {
	log.Info("Starting confd")
	// All flags are defined in the confd/config package which allow us to
	// override configuration settings from the cli. Parse the flags now to
	// make them active.
	flag.Parse()
	if err := InitConfig(); err != nil {
		log.Fatal(err.Error())
	}
	for {
		if err := ProcessTemplateResources(); err != nil {
			log.Error(err.Error())
		}
		// If the -onetime flag is passed on the command line we immediately exit
		// after processing the template config files.
		if Onetime() {
			break
		}
		// By default we poll etcd every 30 seconds
		time.Sleep(time.Duration(Interval()) * time.Second)
	}
}
