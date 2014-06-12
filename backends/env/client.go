package env

import (
	"os"
	"strings"
)

var replacer = strings.NewReplacer("/", "_")

// Client provides a wrapper around the consulkv client
type Client struct{}

// NewEnvClient returns a new client
func NewEnvClient() (*Client, error) {
	return &Client{}, nil
}

// GetValues queries the environment for keys
func (c *Client) GetValues(keys []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	for _, key := range keys {
		k := transform(key)
		value := os.Getenv(k)
		if value != "" {
			vars[key] = value
		}
	}
	return vars, nil
}

func transform(key string) string {
	k := strings.TrimPrefix(key, "/")
	return strings.ToUpper(replacer.Replace(k))
}
