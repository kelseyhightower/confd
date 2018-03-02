package template

import (
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
			log.Error("%s", err.Error())
			lastErr = err
		}
	}
	return lastErr
}

type intervalProcessor struct {
	config   Config
	doneChan chan bool
	errChan  chan error
	interval int
}

func IntervalProcessor(config Config, doneChan chan bool, errChan chan error, interval int) Processor {
	return &intervalProcessor{config, doneChan, errChan, interval}
}

func (p *intervalProcessor) Process() {
	defer close(p.doneChan)
	for {
		ts, err := getTemplateResources(p.config)
		if err != nil {
			p.errChan <- err
			break
		}
		process(ts)
		select {
		case <-time.After(time.Duration(p.interval) * time.Second):
			continue
		}
	}
}

type watchProcessor struct {
	config   Config
	doneChan chan bool
	errChan  chan error
	wg       *sync.WaitGroup
}

func WatchProcessor(config Config, doneChan chan bool, errChan chan error) Processor {
	return &watchProcessor{config, doneChan, errChan, &sync.WaitGroup{}}
}

func (p *watchProcessor) Process() {
	defer close(p.doneChan)
	ts, err := getTemplateResources(p.config)
	if err != nil {
		p.errChan <- err
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
	// Initial template rendering
	if err := t.process(); err != nil {
		p.errChan <- err
	}
	// Waiting for updates
	results := make(chan string)
	defer p.wg.Done()
	keys := appendPrefix(t.Prefix, t.Keys)
	go func() {
		needsUpdate := false
		for {
			select {
			case <-time.Tick(time.Second * 1):
				if needsUpdate {
					needsUpdate = false
					err := t.process()
					if err != nil {
						p.errChan <- err
					}
				}
			case <-results:
				needsUpdate = true
				log.Debug("Got something from the plugin")
			}
		}
	}()
	err := t.database.WatchPrefix(t.Prefix, keys, results)
	if err != nil {
		p.errChan <- err
	}
	return
}

func getTemplateResources(config Config) ([]*TemplateResource, error) {
	var lastError error
	templates := make([]*TemplateResource, 0)
	log.Debug("Loading template resources from confdir " + config.ConfDir)
	if !isFileExist(config.ConfDir) {
		log.Warning("Cannot load template resources: confdir '%s' does not exist", config.ConfDir)
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
		log.Debug("Found template: %s", p)
		t, err := NewTemplateResource(p, config)
		if err != nil {
			lastError = err
			continue
		}
		templates = append(templates, t)
	}
	return templates, lastError
}
