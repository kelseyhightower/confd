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
	quiet             bool
)

func init() {
	flag.StringVar(&configFile, "C", "", "confd config file")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
	flag.BoolVar(&quiet, "q", false, "silence non-error messages")
}

func main() {
	// Most flags are defined in the confd/config package which allows us to
	// override configuration settings from the command line. Parse the flags now
	// to make them active.
	flag.Parse()
	// Non-error messages are not printed by default, enable them now.
	// If the "-q" flag was passed on the command line non-error messages will
	// not be printed.
	log.SetQuiet(quiet)
	log.Info("Starting confd")
	if configFile == "" {
		if IsFileExist(defaultConfigFile) {
			configFile = defaultConfigFile
		}
	}
	// Initialize the global configuration.
	if err := config.LoadConfig(configFile); err != nil {
		log.Fatal(err.Error())
	}
	// Create the etcd client upfront and use it for the life of the process.
	// The etcdClient is an http.Client and designed to be reused.
	log.Debug("Connecting to " + strings.Join(config.EtcdNodes(), ", "))
	etcdClient, err := etcdutil.NewEtcdClient(config.EtcdNodes(), config.ClientCert(), config.ClientKey())
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
