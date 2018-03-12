package file

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/confd/log"
	util "github.com/kelseyhightower/confd/util"
	"gopkg.in/yaml.v2"
)

var replacer = strings.NewReplacer("/", "_")

// Client provides a shell for the yaml client
type Client struct {
	filepath []string
	filter   string
}

type ResultError struct {
	response uint64
	err      error
}

func NewFileClient(filepath []string, filter string) (*Client, error) {
	return &Client{filepath: filepath, filter: filter}, nil
}

func readFile(path string, vars map[string]string) error {
	yamlMap := make(map[interface{}]interface{})
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &yamlMap)
	if err != nil {
		return err
	}

	err = nodeWalk(yamlMap, "/", vars)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	var filePaths []string
	for _, path := range c.filepath {
		p, err := util.RecursiveFilesLookup(path, c.filter)
		if err != nil {
			return nil, err
		}
		filePaths = append(filePaths, p...)
	}

	for _, path := range filePaths {
		err := readFile(path, vars)
		if err != nil {
			return nil, err
		}
	}

VarsLoop:
	for k, _ := range vars {
		for _, key := range keys {
			if strings.HasPrefix(k, key) {
				continue VarsLoop
			}
		}
		delete(vars, k)
	}
	log.Debug(fmt.Sprintf("Key Map: %#v", vars))
	return vars, nil
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
	case int:
		vars[key] = strconv.Itoa(node.(int))
	case bool:
		vars[key] = strconv.FormatBool(node.(bool))
	case float64:
		vars[key] = strconv.FormatFloat(node.(float64), 'f', -1, 64)
	}
	return nil
}

func (c *Client) watchChanges(watcher *fsnotify.Watcher, stopChan chan bool) ResultError {
	outputChannel := make(chan ResultError)
	defer close(outputChannel)
	go func() error {
		for {
			select {
			case event := <-watcher.Events:
				log.Debug("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Create == fsnotify.Create {
					outputChannel <- ResultError{response: 1, err: nil}
				}
			case err := <-watcher.Errors:
				outputChannel <- ResultError{response: 0, err: err}
			case <-stopChan:
				outputChannel <- ResultError{response: 1, err: nil}
			}
		}
	}()
	return <-outputChannel
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
	for _, path := range c.filepath {
		isDir, err := util.IsDirectory(path)
		if err != nil {
			return 0, err
		}
		if isDir {
			dirs, err := util.RecursiveDirsLookup(path, "*")
			if err != nil {
				return 0, err
			}
			for _, dir := range dirs {
				err = watcher.Add(dir)
				if err != nil {
					return 0, err
				}
			}
		} else {
			err = watcher.Add(path)
			if err != nil {
				return 0, err
			}
		}
	}
	output := c.watchChanges(watcher, stopChan)
	if output.response != 2 {
		return output.response, output.err
	}
	return waitIndex, nil
}
