package file

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/confd/log"
	"gopkg.in/yaml.v2"
)

var replacer = strings.NewReplacer("/", "_")

// Client provides a shell for the yaml client
type Client struct {
	filepath string
}

func NewFileClient(filepath string) (*Client, error) {
	return &Client{filepath}, nil
}

func (c *Client) GetValues(keys []string, token string) (map[string]string, error) {
	yamlMap := make(map[interface{}]interface{})
	vars := make(map[string]string)

	data, err := ioutil.ReadFile(c.filepath)
	if err != nil {
		return vars, err
	}
	err = yaml.Unmarshal(data, &yamlMap)
	if err != nil {
		return vars, err
	}

	nodeWalk(yamlMap, "", vars)
	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node map[interface{}]interface{}, key string, vars map[string]string) error {
	for k, v := range node {
		key := key + "/" + k.(string)

		switch v.(type) {
		case map[interface{}]interface{}:
			nodeWalk(v.(map[interface{}]interface{}), key, vars)
		case []interface{}:
			for _, j := range v.([]interface{}) {
				switch j.(type) {
				case map[interface{}]interface{}:
					nodeWalk(j.(map[interface{}]interface{}), key, vars)
				case string:
					vars[key+"/"+j.(string)] = ""
				}
			}
		case string:
			vars[key] = v.(string)
		}
	}
	return nil
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

	err = watcher.Add(c.filepath)
	if err != nil {
		return 0, err
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
