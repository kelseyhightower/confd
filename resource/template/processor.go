package template

import (
	"log"
	"sync"
	"time"
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
			log.Printf("[ERROR] %s", err.Error())
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
			log.Fatal(err.Error())
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
	// Initial template rendering
	if err := t.process(); err != nil {
		p.errChan <- err
	}
	// Waiting for updates
	stream := make(chan error, 10)
	defer p.wg.Done()
	keys := appendPrefix(t.Prefix, t.Keys)
	go func() {
		log.Printf("[DEBUG] Start watching prefix")
		err := t.database.WatchPrefix(t.Prefix, keys, stream)
		if err != nil {
			p.errChan <- err
		}
		log.Printf("[DEBUG] Stop watching prefix")
	}()
	needsUpdate := false
	for {
		select {
		case <-time.After(time.Second * 5):
			if needsUpdate {
				needsUpdate = false
				err := t.process()
				if err != nil {
					p.errChan <- err
				}
			}
		case err := <-stream:
			needsUpdate = true
			if err != nil {
				// Prevent backend errors from consuming all resources.
				p.errChan <- err
				log.Printf("[DEBUG] Sleeping for 2 seconds after backend error")
				time.Sleep(time.Second * 2)
				continue
			}
		}
	}
}

func getTemplateResources(config Config) ([]*TemplateResource, error) {
	var lastError error
	templates := make([]*TemplateResource, 0)
	log.Printf("[DEBUG] Loading template resources from confdir " + config.ConfDir)
	if !isFileExist(config.ConfDir) {
		log.Printf("[WARN] Cannot load template resources: confdir '%s' does not exist", config.ConfDir)
		return nil, nil
	}
	paths, err := recursiveFindFiles(config.ConfigDir, "*toml")
	if err != nil {
		return nil, err
	}

	if len(paths) < 1 {
		log.Printf("[WARN] Found no templates")
	}

	for _, p := range paths {
		log.Printf("[DEBUG] Found template: %s", p)
		t, err := NewTemplateResource(p, config)
		if err != nil {
			lastError = err
			continue
		}
		templates = append(templates, t)
	}
	return templates, lastError
}
