package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
)

func main() {
	flag.Parse()
	if printVersion {
		log.Info("confd %s", version)
		os.Exit(0)
	}
	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Starting confd")

	database, err := backends.New(backendsConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	templateConfig.Database = database
	if onetime {
		if err := template.Process(templateConfig); err != nil {
			log.Fatal(err.Error())
		}
		os.Exit(0)
	}

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	var processor template.Processor
	switch {
	case config.Watch:
		processor = template.WatchProcessor(templateConfig, stopChan, doneChan, errChan)
	default:
		processor = template.IntervalProcessor(templateConfig, stopChan, doneChan, errChan, config.Interval)
	}

	go processor.Process()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			log.Error(err.Error())
		case s := <-signalChan:
			log.Info("Captured %v. Exiting...", s)
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}
