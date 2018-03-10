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

func filesLookup(paths []string) ([]string, error) {
	var filePaths []string
	for _, path := range paths {
		f, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		switch mode := f.Mode(); {
		case mode.IsDir():
			fileList := make([]string, 0)
			e := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
				fileList = append(fileList, path)
				return err
			})
			if e != nil {
				return nil, e
			}
			filePaths = append(filePaths, fileList...)

		case mode.IsRegular():
			filePaths = append(filePaths, path)
		}
	}
	return filePaths, nil
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

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	if waitIndex == 0 {
		return 1, nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return 0, err
	}
	defer watcher.Close()

	filePaths, err := filesLookup(c.filepath)
	if err != nil {
		return 0, err
	}

	for _, path := range filePaths {
		err = watcher.Add(path)
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
