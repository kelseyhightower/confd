package file

import (
	"fmt"
	"io/ioutil"
	"strings"
	"path"
	"errors"
	"strconv"

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
			if key != "" {
				filteredYamlMap, err := filterByPath(key, yamlMap)
				if err != nil {
					return vars, err
				}
				nodeWalk(filteredYamlMap, key, vars)
			}
		}
	}

	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

func filterByPath(path string, varsMap interface{}) (interface{}, error) {
	keys := strings.Split(path, "/")
	filteredVarsMap := varsMap
	for _, key := range(keys) {
		if key != "" {
			switch filteredVarsMap.(type) {
				case map[interface{}]interface{}:
					newFilteredVarsMap, exists := filteredVarsMap.(map[interface{}]interface{})[key]
					if !exists {
						message := fmt.Sprintf("Error: cannot find element in %v with a key %s", varsMap, key)
						return nil, errors.New(message)
					}
					filteredVarsMap = newFilteredVarsMap
				case []interface{}:
					index, err := strconv.Atoi(key)
					if err != nil {
						return nil, err
					}
					newFilteredVarsMap := filteredVarsMap.([]interface{})[index]
					filteredVarsMap = newFilteredVarsMap
				default:
					message := fmt.Sprintf("Error: element %v has wrong type. Map or slice is expected!", filteredVarsMap)
					return nil, errors.New(message)
			}
		}
	}
	return filteredVarsMap, nil
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node interface{}, key string, vars map[string]string) error {
	switch node.(type) {
		case []interface{}:
			for i, j := range node.([]interface{}) {
				key := path.Join(key, strconv.Itoa(i))
				nodeWalk(j, key, vars)
			}
		case map[interface{}]interface{}:
			for k, v := range node.(map[interface{}]interface{}) {
				key := path.Join(key, k.(string))
				nodeWalk(v, key, vars)
			}
		case string:
			vars[key] = node.(string)
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
