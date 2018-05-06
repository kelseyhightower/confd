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
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

var replacer = strings.NewReplacer("/", "_")

// Client provides a shell for the yaml client
type Client struct {
	filepath []string
	filter   string
}

func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}
	c.filepath = strings.Split(config.YamlFile, ",")
	c.filter = config.Filter
	return nil
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

func (c *Client) WatchPrefix(prefix string, keys []string, results chan string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	for _, path := range c.filepath {
		isDir, err := util.IsDirectory(path)
		if err != nil {
			return err
		}
		if isDir {
			dirs, err := util.RecursiveDirsLookup(path, "*")
			if err != nil {
				return err
			}
			for _, dir := range dirs {
				err = watcher.Add(dir)
				if err != nil {
					return err
				}
			}
		} else {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
	}

	for {
		select {
		case event := <-watcher.Events:
			log.Debug("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Create == fsnotify.Create {
				results <- ""
			}
		case err := <-watcher.Errors:
			log.Error(err.Error())
		}
	}
}
