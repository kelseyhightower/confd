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
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	if len(os.Args) > 1 && os.Args[1] == "internal-plugin" {
		log.Info("Plugin is about to start")
		exitCode := backends.RunPlugin(os.Args[2:])
		log.Info("Plugin is about to exit with %#v exit code", exitCode)
		return exitCode
	}
	flag.Parse()
	if printVersion {
		log.Info("confd %s", version)
		return 0
	}
	if err := initConfig(); err != nil {
		log.Error(err.Error())
		return 1
	}

	log.Info("Starting confd")

	database, client, err := backends.New(backendsConfig)
	defer client.Kill()
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	templateConfig.Database = database
	if onetime {
		if err := template.Process(templateConfig); err != nil {
			log.Error(err.Error())
			return 1
		}
		return 0
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
			return 0
		}
	}
}
