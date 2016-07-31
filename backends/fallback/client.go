package fallback

import (
        "fmt"

        "github.com/kelseyhightower/confd/log"
)


type StoreClient interface {
        GetValues(keys []string) (map[string]string, error)
        WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

// Client provides a shell for the env client
type FallbackClient struct{
	mainBackend StoreClient
	fallbackBackend StoreClient
}

// NewEnvClient returns a new client
func NewFallbackClient(mainBackend StoreClient, fallbackBackend StoreClient) (*FallbackClient, error) {
        return &FallbackClient{mainBackend, fallbackBackend}, nil
}

// GetValues queries the environment for keys
func (c *FallbackClient) GetValues(keys []string) (map[string]string, error) {
        fallback_vars, err := c.fallbackBackend.GetValues(keys)
        if err != nil {
                return nil, err
        }

	vars, err := c.mainBackend.GetValues(keys)
	if err == nil {
		log.Info(fmt.Sprintf("Key Map: %#v", vars))
		for k, v := range vars {
			fallback_vars[k] = v
		}
	}

        log.Info(fmt.Sprintf("Key Map: %#v", fallback_vars))
        return fallback_vars, nil
}


func (c *FallbackClient) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
        return c.mainBackend.WatchPrefix(prefix, keys, waitIndex, stopChan)
}

