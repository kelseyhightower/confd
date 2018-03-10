package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/confd/log"
	"gopkg.in/yaml.v2"
)

var replacer = strings.NewReplacer("/", "_")

// Client provides a shell for the yaml client
type Client struct {
	filepath []string
}

type ResultError struct {
	response uint64
	err      error
}

func NewFileClient(filepath []string) (*Client, error) {
	return &Client{filepath}, nil
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

func isDirectory(path string) (bool, error) {
	f, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	switch mode := f.Mode(); {
	case mode.IsDir():
		return true, nil
	case mode.IsRegular():
		return false, nil
	}
	return false, nil
}

func filesLookup(paths []string) ([]string, error) {
	var files []string
	for _, path := range paths {
		isDir, err := isDirectory(path)
		if err != nil {
			return nil, err
		}
		if isDir {
			err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
				isDir, err := isDirectory(path)
				if err != nil {
					return err
				}
				if !isDir {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			files = append(files, path)
		}
	}
	return files, nil
}

func watcherTargetLookup(paths []string, watcher *fsnotify.Watcher) error {
	for _, path := range paths {
		isDir, err := isDirectory(path)
		if err != nil {
			return err
		}
		if isDir {
			err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
				isDir, err := isDirectory(path)
				if err != nil {
					return err
				}
				if isDir {
					err = watcher.Add(path)
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	filePaths, err := filesLookup(c.filepath)
	if err != nil {
		return nil, err
	}

	for _, path := range filePaths {
		readFile(path, vars)
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
	go func() error {
		for {
			select {
			case event := <-watcher.Events:
				log.Debug("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Rename == fsnotify.Rename ||
					event.Op&fsnotify.Chmod == fsnotify.Chmod {
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

	err = watcherTargetLookup(c.filepath, watcher)
	if err != nil {
		return 0, err
	}
	output := c.watchChanges(watcher, stopChan)
	if output.response != 2 {
		return output.response, output.err
	}
	return waitIndex, nil
}
