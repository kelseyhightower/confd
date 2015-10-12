package template

import (
	"sync"
	"time"

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
)

type Processor interface {
	Process()
}

func Process(tcs []*config.TemplateConfig, storeClient backends.StoreClient, noop bool) error {
	return process(getTemplateResources(tcs, storeClient), noop)
}

type intervalProcessor struct {
	trs         []*TemplateResource
	storeClient backends.StoreClient
	noop        bool

	stopChan    chan bool
	doneChan    chan bool
	errChan     chan error
	interval    int
}

func IntervalProcessor(tcs []*config.TemplateConfig, storeClient backends.StoreClient, noop bool,
                       stopChan, doneChan chan bool, errChan chan error, interval int) Processor {
	trs := getTemplateResources(tcs, storeClient)
	return &intervalProcessor{trs, storeClient, noop, stopChan, doneChan, errChan, interval}
}

func (p *intervalProcessor) Process() {
	defer close(p.doneChan)
	for {
		process(p.trs, p.noop)
		select {
		case <-p.stopChan:
			break
		case <-time.After(time.Duration(p.interval) * time.Second):
			continue
		}
	}
}

type watchProcessor struct {
	trs         []*TemplateResource
	storeClient backends.StoreClient
	noop        bool

	stopChan    chan bool
	doneChan    chan bool
	errChan     chan error
	wg          sync.WaitGroup
}

func WatchProcessor(tcs []*config.TemplateConfig, storeClient backends.StoreClient, noop bool,
                    stopChan, doneChan chan bool, errChan chan error) Processor {
	var wg sync.WaitGroup
	trs := getTemplateResources(tcs, storeClient)
	return &watchProcessor{trs, storeClient, noop, stopChan, doneChan, errChan, wg}
}

func (p *watchProcessor) Process() {
	defer close(p.doneChan)
	for _, t := range p.trs {
		t := t
		p.wg.Add(1)
		go p.monitorPrefix(t)
	}
	p.wg.Wait()
}

func (p *watchProcessor) monitorPrefix(t *TemplateResource) {
	defer p.wg.Done()
	for {
		index, err := t.storeClient.WatchPrefix(t.config.Prefix, t.lastIndex, p.stopChan)
		if err != nil {
			p.errChan <- err
			// Prevent backend errors from consuming all resources.
			time.Sleep(time.Second * 2)
			continue
		}
		t.lastIndex = index
		if err := t.process(p.noop); err != nil {
			p.errChan <- err
		}
	}
}

func process(trs []*TemplateResource, noop bool) error {
	var lastErr error
	for _, t := range trs {
		if err := t.process(noop); err != nil {
			log.Error(err.Error())
			lastErr = err
		}
	}
	return lastErr
}

func getTemplateResources(tcs []*config.TemplateConfig, storeClient backends.StoreClient) []*TemplateResource {
	trs := make([]*TemplateResource, 0)
	for _, tc := range tcs {
		trs = append(trs, NewTemplateResource(tc, storeClient))
	}
	return trs
}