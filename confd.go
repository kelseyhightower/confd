// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"flag"
	"os"
	"time"

	"github.com/kelseyhightower/confd/log"
)

var (
	configFile        = ""
	defaultConfigFile = "/etc/confd/confd.toml"
	onetime           bool
	quiet             bool
)

func init() {
	flag.StringVar(&configFile, "C", "", "confd config file")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
	flag.BoolVar(&quiet, "q", false, "silence non-error messages")
}

func main() {
	// Most flags are defined in the confd/config package which allow us to
	// override configuration settings from the cli. Parse the flags now to
	// make them active.
	flag.Parse()
	// non-error messages are not printed by default, enable them now.
	// If the "-q" flag was passed on the commandline non-error messages will
	// not be printed.
	log.SetQuiet(quiet)
	log.Info("Starting confd")
	if configFile == "" {
		if IsFileExist(defaultConfigFile) {
			configFile = defaultConfigFile
		}
	}
	if err := loadConfig(configFile); err != nil {
		log.Fatal(err.Error())
	}
	// Create the etcd client upfront and use it for the life of the process.
	// The etcdClient is an http.Client and designed to be reused.
	etcdClient, err := newEtcdClient(EtcdNodes(), ClientCert(), ClientKey())
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		runErrors := ProcessTemplateResources(etcdClient)
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
