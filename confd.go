// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
)

var (
	configFile        = ""
	defaultConfigFile = "/etc/confd/confd.toml"
	onetime           bool
	printVersion      bool
)

func init() {
	flag.StringVar(&configFile, "config-file", "", "the confd config file")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
	flag.BoolVar(&printVersion, "version", false, "print the version and exit")
}

func main() {
	// Most flags are defined in the confd/config package which allows us to
	// override configuration settings from the command line. Parse the flags now
	// to make them active.
	flag.Parse()
	if printVersion {
		fmt.Printf("confd %s\n", Version)
		os.Exit(0)
	}
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

	// Create the storage client
	log.Notice("Backend set to " + config.Backend())
	store, err := backends.New(config.Backend())
	if err != nil {
		log.Fatal(err.Error())
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		runErrors := template.ProcessTemplateResources(store)
		// If the -onetime flag is passed on the command line we immediately exit
		// after processing the template config files.
		if onetime {
			if len(runErrors) > 0 {
				os.Exit(1)
			}
			os.Exit(0)
		}
		select {
		case c := <-signalChan:
			log.Info(fmt.Sprintf("captured %v exiting...", c))
			os.Exit(0)
		case <-time.After(time.Duration(config.Interval()) * time.Second):
			// Continue processing templates.
		}
	}
}

// IsFileExist reports whether path exits.
func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
