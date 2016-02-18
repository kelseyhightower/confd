package template

import (
	"fmt"
	"strings"
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
	for {
		index, err := t.storeClient.WatchPrefix(t.Prefix, t.lastIndex, p.stopChan)
		if err != nil {
			p.errChan <- err
			// Prevent backend errors from consuming all resources.
			time.Sleep(time.Second * 2)
			continue
		}
		t.lastIndex = index
		if err := t.process(); err != nil {
			p.errChan <- err
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

	// get dynamic TemplateResources
	log.Info("parse dynamic keys")
	result, err := config.StoreClient.GetValues([]string{"_confd"})
	if err == nil {
		for k, v := range result {
			log.Info("dynamic key: " + k + " / " + v)
			t, err := NewTemplateResource(config.ConfigDir+"/"+v, config)
			if err != nil {
				lastError = err
				continue
			}

			split := strings.Split(k, "/")
			key := "/" + split[len(split)-1]
			t.Dest = strings.Replace(t.Dest, "{{.token}}", key, 1)
			t.ReloadCmd = strings.Replace(t.ReloadCmd, "{{.token}}", key, 1)
			t.Prefix = key
			t.prefix = key
			templates = append(templates, t)
		}
	}

	return templates, lastError
}
