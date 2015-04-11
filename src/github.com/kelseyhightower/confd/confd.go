package confd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
	"github.com/kelseyhightower/confd/util"
)

func Run(gc *config.GlobalConfig, tcs []*config.TemplateConfig, bc config.BackendConfig) {
	// Configure logging.
	if gc.LogLevel != "" {
		log.SetLevel(gc.LogLevel)
	}

	// Dump input parameters, just for debugging purposes
	util.Dump(gc)
	for _, tc := range tcs {
		util.Dump(tc)
	}
	util.Dump(bc)

	// prepend global prefix to template prefix (if provided)
	if gc.Prefix != "" {
		for _, tc := range tcs {
			tc.Prefix = filepath.Join("/", gc.Prefix, tc.Prefix)
		}
	}

	// Exit if watch is requested and not supported by backend
	if gc.Watch && !bc.IsWatchSupported() {
		log.Info(fmt.Sprintf("Watch is not supported for backend %s. Exiting...", bc.Type()))
		os.Exit(1)
	}

	// Notify which backend is going to use
	log.Info("Backend set to " + bc.Type())

	// Create store client instance
	storeClient, err := backends.New(bc)
	if err != nil {
		log.Fatal(err.Error())
	}

	// if onetime execution is requested, do it and then exit
	if gc.Onetime {
		if err := template.Process(tcs, storeClient, gc.Noop); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// main loop
	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)
	var processor template.Processor
	switch {
		case gc.Watch:
		processor = template.WatchProcessor(tcs, storeClient, gc.Noop, stopChan, doneChan, errChan)
		break
		default:
		processor = template.IntervalProcessor(tcs, storeClient, gc.Noop, stopChan, doneChan, errChan, gc.Interval)
		break
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