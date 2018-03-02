package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
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
		log.Info("confd %s (Git SHA: %s, Go Version: %s)\n", Version, GitSHA, runtime.Version())
		os.Exit(0)
	}
	if err := initConfig(); err != nil {
		log.Error("%s", err.Error())
		return 1
	}

	log.Info("Starting confd")

	database, client, err := backends.New(backendsConfig)
	if err != nil {
		log.Error("Failed to connect to a plugin. %s", err.Error())
		return 1
	}
	defer func() {
		log.Info("Closing a plugin process")
		client.Kill()
	}()

	templateConfig.Database = database
	if onetime {
		if err := template.Process(templateConfig); err != nil {
			log.Error("Failed to process a template. %s", err.Error())
			return 1
		}
		return 0
	}

	doneChan := make(chan bool)
	errChan := make(chan error)

	go func(errChan chan error) {
		for err := range errChan {
			log.Error("Received an error from a plugin. %s", err.Error())
		}
	}(errChan)

	var processor template.Processor
	switch {
	case config.Watch:
		processor = template.WatchProcessor(templateConfig, doneChan, errChan)
	default:
		processor = template.IntervalProcessor(templateConfig, doneChan, errChan, config.Interval)
	}

	go processor.Process()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-signalChan:
		log.Info("Captured %s. Exiting...", s)
		return 0
	case <-doneChan:
		log.Info("Exiting...")
		return 0
	}
}
