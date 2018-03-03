package file

import (
	"fmt"
	"io/ioutil"
	"strings"
	"path"
	"errors"

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

func (c *Client) GetValues(keys []string) (map[string]string, error) {
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

	if len(keys) == 0 {
		nodeWalk(yamlMap, "", vars)
	} else {
		for _, key := range keys {
			filteredYamlMap, err := filterByPath(key, yamlMap)
			if err != nil {
				return vars, err
			}
			nodeWalk(filteredYamlMap, key, vars)
		}
	}

	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

func filterByPath(path string, varsMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	keys := strings.Split(path, "/")
	filteredVarsMap := varsMap
	for _, key := range(keys) {
		if key != "" {
			newFilteredVarsMap, err := filterByKey(key, filteredVarsMap)
			if err != nil {
				return varsMap, err
			}
			filteredVarsMap = newFilteredVarsMap
		}
	}
	return filteredVarsMap, nil
}

func filterByKey(key string, varsMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	newVarsMap, exists := varsMap[key].(map[interface{}]interface{})
	if !exists {
		message := fmt.Sprintf("Error: cannot find element in %v with a key %s", varsMap, key)
		return make(map[interface{}]interface{}), errors.New(message)
	}
	return newVarsMap, nil
}


// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node map[interface{}]interface{}, key string, vars map[string]string) error {
	for k, v := range node {
		key := path.Join(key, k.(string))
		switch v.(type) {
		case map[interface{}]interface{}:
			nodeWalk(v.(map[interface{}]interface{}), key, vars)
		case []interface{}:
			for _, j := range v.([]interface{}) {
				switch j.(type) {
				case map[interface{}]interface{}:
					nodeWalk(j.(map[interface{}]interface{}), key, vars)
				case string:
					vars[path.Join(key, j.(string))] = ""
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
