package dotenv

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Client provides a shell for the env client
type Client struct {
	filePath string
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

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
