package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/confd"
	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
)

func main() {
	flag.Parse()
	if confd.PrintVersion {
		fmt.Printf("confd %s\n", confd.Version)
		os.Exit(0)
	}
	if err := confd.InitConfig(); err != nil {
		log.Fatal(err.Error())
	}
	log.Info("Starting confd")
	storeClient, err := backends.New(confd.BackendsConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	confd.TemplateConfig.StoreClient = storeClient
	if confd.Onetime {
		if err := template.Process(confd.TemplateConfig); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)
	var processor template.Processor
	switch {
	case confd.Cfg.Watch:
		processor = template.WatchProcessor(confd.TemplateConfig, stopChan, doneChan, errChan)
	default:
		processor = template.IntervalProcessor(confd.TemplateConfig, stopChan, doneChan, errChan, confd.Cfg.Interval)
	}
	go processor.Process()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			log.Error(err.Error())
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}
