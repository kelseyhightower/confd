package env

import (
	"errors"
	"os"
	"strings"

	"github.com/kelseyhightower/confd/builtin/databases/env/util"
	"github.com/kelseyhightower/confd/log"
)

// Client provides a shell for the env client
type Client struct{}

func (c *Client) Configure(map[string]string) error {
	return nil
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
		k := util.Transform(key)
		for envKey, envValue := range envMap {
			if strings.HasPrefix(envKey, k) {
				vars[util.Clean(envKey)] = envValue
			}
		}
	}

	log.Debug("Key Map: %#v", vars)

	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, results chan string) error {
	return errors.New("WatchPrefix is not implemented")
}
