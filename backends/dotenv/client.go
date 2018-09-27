package dotenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/confd/log"
)

// Client provides a shell for the env client
type Client struct {
	filePath string
}

type ResultError struct {
	response uint64
	err      error
}

// NewDotenvClient returns a new client
func NewDotenvClient(dotenvFile string) (*Client, error) {
	return &Client{filePath: dotenvFile}, nil
}

// GetValues queries the environment for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	// Load .env values
	var err error
	if c.filePath != "" {
		err = godotenv.Load(c.filePath)
	} else {
		err = godotenv.Load()
	}
	if err != nil {
		return nil, err
	}

	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		index := strings.Index(e, "=")
		envMap[e[:index]] = e[index+1:]
	}

	vars := make(map[string]string)
	for _, key := range keys {
		for envKey, envValue := range envMap {
			if strings.HasPrefix(envKey, key) {
				vars[envKey] = envValue
			}
		}
	}

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

	watcher.Add(c.filePath)

	output := c.watchChanges(watcher, stopChan)
	if output.response != 2 {
		return output.response, output.err
	}
	return waitIndex, nil
}

func (c *Client) watchChanges(watcher *fsnotify.Watcher, stopChan chan bool) ResultError {
	outputChannel := make(chan ResultError)
	go func() error {
		defer close(outputChannel)
		for {
			select {
			case event := <-watcher.Events:
				log.Debug(fmt.Sprintf("Event: %s", event))
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
