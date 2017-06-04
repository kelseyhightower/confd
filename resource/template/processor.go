package template

import (
	"fmt"
	"sync"
	"time"

	"github.com/kelseyhightower/confd/log"
)

type Processor interface {
	Process()
}

func Process(config Config) error {
	ts, err := getTemplateResources(config)


	if err != nil {
		return err
	}

	return process(ts)
}

func process(ts []*TemplateResource) error {
	var lastErr error
	for _, t := range ts {
		if err := t.process(); err != nil {
			log.Error(err.Error())
			lastErr = err
		}
	}
	return lastErr
}

type intervalProcessor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	interval int
}

func IntervalProcessor(config Config, stopChan, doneChan chan bool, errChan chan error, interval int) Processor {
	return &intervalProcessor{config, stopChan, doneChan, errChan, interval}
}

func (p *intervalProcessor) Process() {
	defer close(p.doneChan)
	for {
		ts, err := getTemplateResources(p.config)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
		process(ts)
		select {
		case <-p.stopChan:
			break
		case <-time.After(time.Duration(p.interval) * time.Second):
			continue
		}
	}
}

type watchProcessor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	wg       sync.WaitGroup
}

func WatchProcessor(config Config, stopChan, doneChan chan bool, errChan chan error) Processor {
	var wg sync.WaitGroup
	return &watchProcessor{config, stopChan, doneChan, errChan, wg}
}

func (p *watchProcessor) Process() {
	defer close(p.doneChan)
	ts, err := getTemplateResources(p.config)

	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for _, t := range ts {
		t := t
		p.wg.Add(1)
		go p.monitorPrefix(t)
	}
	p.wg.Wait()
}

func (p *watchProcessor) monitorPrefix(t *TemplateResource) {
	defer p.wg.Done()

	var keys2 []string

	continueLoop := make(chan bool)

	keys := appendPrefix(t.Prefix, t.Keys)

	// if second backend is enabled
	if t.storeClient2 != nil {
		keys2 = appendPrefix(t.Prefix2, t.Keys2)
	}

	for {


		go func(p *watchProcessor,t *TemplateResource) {
			index, err := t.storeClient.WatchPrefix(t.Prefix, keys, t.lastIndex, p.stopChan)
			if err != nil {
				p.errChan <- err
				// Prevent backend errors from consuming all resources.
				time.Sleep(time.Second * 2)

				return
			}
			t.lastIndex = index

			continueLoop<-true
		}(p,t)

		// if second backend is enabled
		if t.storeClient2 != nil {

			go func(p *watchProcessor, t *TemplateResource) {
				log.Debug("b2 Watch Step2 iteration start ")
				index, err := t.storeClient2.WatchPrefix(t.Prefix2, keys2, t.lastIndex2, p.stopChan)
				log.Debug("b2 Watch Step2 iteration ")
				if err != nil {
					p.errChan <- err
					// Prevent backend errors from consuming all resources.
					time.Sleep(time.Second * 2)

					return
				}
				t.lastIndex2 = index
				continueLoop <- true
			}(p, t)

		}

		select {
		case <-continueLoop:

			if err := t.process(); err != nil {
				p.errChan <- err
			}
			continue
		}

	}

}

func getTemplateResources(config Config) ([]*TemplateResource, error) {
	var lastError error

	templates := make([]*TemplateResource, 0)
	log.Debug("Loading template resources from confdir " + config.ConfDir)
	if !isFileExist(config.ConfDir) {
		log.Warning(fmt.Sprintf("Cannot load template resources: confdir '%s' does not exist", config.ConfDir))
		return nil, nil
	}
	paths, err := recursiveFindFiles(config.ConfigDir, "*toml")
	if err != nil {
		return nil, err
	}

	if len(paths) < 1 {
		log.Warning("Found no templates")
	}

	for _, p := range paths {
		log.Debug(fmt.Sprintf("Found template: %s", p))
		t, err := NewTemplateResource(p, config)
		if err != nil {
			lastError = err
			continue
		}
		templates = append(templates, t)
	}
	return templates, lastError
}
