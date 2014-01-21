// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/etcd/etcdutil"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
)

var (
	configFile        = ""
	defaultConfigFile = "/etc/confd/confd.toml"
	onetime           bool
)

func init() {
	flag.StringVar(&configFile, "config-file", "", "the confd config file")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
}

func main() {
	// Most flags are defined in the confd/config package which allows us to
	// override configuration settings from the command line. Parse the flags now
	// to make them active.
	flag.Parse()
	if configFile == "" {
		if IsFileExist(defaultConfigFile) {
			configFile = defaultConfigFile
		}
	}
	// Initialize the global configuration.
	log.Debug("Loading confd configuration")
	if err := config.LoadConfig(configFile); err != nil {
		log.Fatal(err.Error())
	}
	// Configure logging. While you can enable debug and verbose logging, however
	// if quiet is set to true then debug and verbose messages will not be printed.
	log.SetQuiet(config.Quiet())
	log.SetVerbose(config.Verbose())
	log.SetDebug(config.Debug())
	log.Notice("Starting confd")
	// Create the etcd client upfront and use it for the life of the process.
	// The etcdClient is an http.Client and designed to be reused.
	log.Notice("etcd nodes set to " + strings.Join(config.EtcdNodes(), ", "))
	etcdClient, err := etcdutil.NewEtcdClient(config.EtcdNodes(), config.ClientCert(), config.ClientKey(), config.ClientCaKeys())
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		runErrors := template.ProcessTemplateResources(etcdClient)
		// If the -onetime flag is passed on the command line we immediately exit
		// after processing the template config files.
		if onetime {
			if len(runErrors) > 0 {
				os.Exit(1)
			}
			os.Exit(0)
		}
		time.Sleep(time.Duration(config.Interval()) * time.Second)
	}
}

// IsFileExist reports whether path exits.
func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
