package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/bacongobbler/confd/log"
)

var replacer = strings.NewReplacer("/", "_")

// Client provides a shell for the env client
type Client struct{
	envSep string
}

// NewEnvClient returns a new client
func NewEnvClient(envSep string) (*Client, error) {
	return &Client{envSep: envSep}, nil
}

// GetValues queries the environment for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	replacer := strings.NewReplacer("/", c.envSep)
	allEnvVars := os.Environ()
	envMap := make(map[string]string)
	for _, e := range allEnvVars {
		index := strings.Index(e, "=")
		envMap[e[:index]] = e[index+1:]
	}
	cleanReplacer := strings.NewReplacer(c.envSep, "/")
	vars := make(map[string]string)
	for _, key := range keys {
		k := transform(key, replacer)
		for envKey, envValue := range envMap {
			if strings.HasPrefix(envKey, k) {
				vars[clean(envKey, cleanReplacer)] = envValue
			}
		}
	}

	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

func transform(key string, replacer *strings.Replacer) string {
	k := strings.TrimPrefix(key, "/")
	return strings.ToUpper(replacer.Replace(k))
}

func clean(key string, replacer *strings.Replacer) string {
	newKey := "/" + key
	return replacer.Replace(strings.ToLower(newKey))
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
