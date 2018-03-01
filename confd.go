package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/logging"
	"github.com/kelseyhightower/confd/resource/template"
)

func main() {
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	logging.SetLevel("INFO")
	if len(os.Args) > 1 && os.Args[1] == "internal-plugin" {
		log.Printf("[INFO] Plugin is about to start")
		exitCode := backends.RunPlugin(os.Args[2:])
		log.Printf("[INFO] Plugin is about to exit with %#v exit code", exitCode)
		return exitCode
	}
	flag.Parse()
	if printVersion {
		log.Printf("[INFO] confd %s (Git SHA: %s, Go Version: %s)\n", Version, GitSHA, runtime.Version())
		os.Exit(0)
	}
	if err := initConfig(); err != nil {
		log.Printf("[ERROR] %s", err.Error())
		return 1
	}

	log.Printf("[INFO] Starting confd")

	database, client, err := backends.New(backendsConfig)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to a plugin. %s", err.Error())
		return 1
	}
	defer func() {
		log.Printf("[INFO] Closing a plugin process")
		client.Kill()
	}()

	templateConfig.Database = database
	if onetime {
		if err := template.Process(templateConfig); err != nil {
			log.Printf("[ERROR] Failed to process a template. %s", err.Error())
			return 1
		}
		return 0
	}

	doneChan := make(chan bool)
	errChan := make(chan error)

	go func(errChan chan error) {
		for err := range errChan {
			log.Printf("[ERROR] Received an error from a plugin. %s", err.Error())
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
		log.Printf("[INFO] Captured %s. Exiting...", s)
		return 0
	case <-doneChan:
		log.Printf("[INFO] Exiting...")
		return 0
	}
}
