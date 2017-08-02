package env

import (
	"log"
	"os"
	"strings"

	"github.com/kelseyhightower/confd/builtin/databases/env/util"
	"github.com/kelseyhightower/confd/confd"
)

// Client provides a shell for the env client
type Client struct{}

// Database returns a new client
func Database() confd.Database {
	return &Client{}
}

func (c *Client) Configure(map[string]interface{}) error {
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

	log.Printf("Key Map: %#v", vars)

	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64) (uint64, error) {
	return 0, nil
}
