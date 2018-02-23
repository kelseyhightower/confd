package clconf

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/confd/log"
	realclconf "gitlab.com/pastdev/s2i/clconf/clconf"
)

// Client provides a shell for the yaml client
type Client struct {
	yamlFiles []string
}

func NewClconfClient(yamlFiles string) (*Client, error) {
	return &Client{realclconf.Splitter.Split(yamlFiles, -1)}, nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	yamlMap, err := realclconf.LoadConf(c.yamlFiles, []string{})
	if err != nil {
		return vars, err
	}

	vars = realclconf.ToKvMap(yamlMap)
	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	if waitIndex == 0 {
		return 1, nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return 0, err
	}
	defer watcher.Close()

	for _, filepath := range c.yamlFiles {
		err = watcher.Add(filepath)
		if err != nil {
			return 0, err
		}
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Remove == fsnotify.Remove {
				return 1, nil
			}
		case err := <-watcher.Errors:
			return 0, err
		case <-stopChan:
			return 0, nil
		}
	}
	return waitIndex, nil
}
