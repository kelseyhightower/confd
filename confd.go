// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"flag"
	"time"
	"os"

	"github.com/kelseyhightower/confd/log"
)

var (
	configFile        = ""
	defaultConfigFile = "/etc/confd/confd.toml"
	onetime           bool
)

func init() {
	flag.StringVar(&configFile, "C", "", "confd config file")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
}

func main() {
	log.Info("Starting confd")
	// Most flags are defined in the confd/config package which allow us to
	// override configuration settings from the cli. Parse the flags now to
	// make them active.
	flag.Parse()
	if configFile == "" {
		if IsFileExist(defaultConfigFile) {
			configFile = defaultConfigFile
		}
	}
	if err := loadConfig(configFile); err != nil {
		log.Fatal(err.Error())
	}
	for {
		runErrors := make([]error, 0)
		if err := ProcessTemplateResources(nil); err != nil {
			runErrors = append(runErrors, err)
			log.Error(err.Error())
		}
		// If the -onetime flag is passed on the command line we immediately exit
		// after processing the template config files.
		if onetime {
			if len(runErrors) > 0 {
				os.Exit(1)
			}
			os.Exit(0)
		}
		// By default we poll etcd every 30 seconds
		time.Sleep(time.Duration(Interval()) * time.Second)
	}
}
