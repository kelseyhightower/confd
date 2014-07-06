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

	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("confd %s\n", Version)
		os.Exit(0)
	}
	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}
	log.Notice("Starting confd")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		runErrors := template.ProcessTemplateResources(templateConfig)
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
		case <-time.After(time.Duration(config.Interval) * time.Second):
			// Continue processing templates.
		}
	}
}
