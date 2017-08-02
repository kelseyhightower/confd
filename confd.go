package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
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
		log.Printf("[INFO] confd %s", version)
		return 0
	}
	if err := initConfig(); err != nil {
		log.Printf("[ERROR] %s", err.Error())
		return 1
	}

	log.Printf("[INFO] Starting confd")

	database, client, err := backends.New(backendsConfig)
	defer client.Kill()
	if err != nil {
		log.Printf("[ERROR] %s", err.Error())
		return 1
	}

	templateConfig.Database = database
	if onetime {
		if err := template.Process(templateConfig); err != nil {
			log.Printf("[ERROR] %s", err.Error())
			return 1
		}
		return 0
	}

	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	var processor template.Processor
	switch {
	case config.Watch:
		processor = template.WatchProcessor(templateConfig, doneChan, errChan)
	default:
		processor = template.IntervalProcessor(templateConfig, doneChan, errChan, config.Interval)
	}

	go processor.Process()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			log.Printf("[ERROR] %s", err.Error())
		case s := <-signalChan:
			log.Printf("[INFO] Captured %v. Exiting...", s)
			close(doneChan)
		case <-doneChan:
			return 0
		}
	}
}
