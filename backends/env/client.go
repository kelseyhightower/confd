package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/confd/log"
)

var replacer = strings.NewReplacer("/", "_")

// Client provides a shell for the env client
type Client struct{}

// NewEnvClient returns a new client
func NewEnvClient() (*Client, error) {
	return &Client{}, nil
}

// GetValues queries the environment for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	allEnvVars := os.Environ()
	envMap := make(map[string]string)
	for _, e := range allEnvVars {
		index := strings.Index(e, "=")
		envMap[e[:index]] = e[index+1:]
	}
	vars := make(map[string]string)
	for _, key := range keys {
		k := transform(key)
		for envKey, envValue := range envMap {
			if strings.HasPrefix(envKey, k) {
				vars[clean(envKey)] = envValue
			}
		}
	}

	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

func transform(key string) string {
	k := strings.TrimPrefix(key, "/")
	return strings.ToUpper(replacer.Replace(k))
}

var cleanReplacer = strings.NewReplacer("_", "/")

func clean(key string) string {
	newKey := "/" + key
	return cleanReplacer.Replace(strings.ToLower(newKey))
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
