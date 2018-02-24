package clconf

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/confd/log"
	realclconf "github.com/pastdev/clconf/clconf"
)

// Client provides a shell for the yaml client
type Client struct {
	yamlFiles         []string
	yamlBase64Strings []string
}

func NewClconfClient(yamlFiles, yamlBase64Strings string) (*Client, error) {
	var yamlFileArray, yamlBase64StringArray []string
	if yamlFiles != "" {
		yamlFileArray = realclconf.Splitter.Split(yamlFiles, -1)
	}
	if yamlBase64Strings != "" {
		yamlBase64StringArray = realclconf.Splitter.Split(yamlBase64Strings, -1)
	}
	return &Client{yamlFileArray, yamlBase64StringArray}, nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	yamlMap, err := realclconf.LoadConfFromEnvironment(
		c.yamlFiles, c.yamlBase64Strings)
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
