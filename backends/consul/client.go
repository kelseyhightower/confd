package consul

import (
	"path"
	"strings"

	"github.com/armon/consul-api"
)

// Client provides a wrapper around the consulkv client
type Client struct {
	client *consulapi.KV
}

// NewConsulClient returns a new client to Consul for the given address
func NewConsulClient(nodes []string) (*Client, error) {
	conf := consulapi.DefaultConfig()
	if len(nodes) > 0 {
		conf.Address = nodes[0]
	}
	client, err := consulapi.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &Client{client.KV()}, nil
}

// GetValues queries Consul for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		key := strings.TrimPrefix(key, "/")
		pairs, _, err := c.client.List(key, nil)
		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[path.Join("/", p.Key)] = string(p.Value)
		}
	}
	return vars, nil
}

type watchResponse struct {
	waitIndex uint64
	err       error
}

func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	respChan := make(chan watchResponse)
	go func() {
		opts := consulapi.QueryOptions{
			WaitIndex: waitIndex,
		}
		_, meta, err := c.client.List(prefix, &opts)
		respChan <- watchResponse{meta.LastIndex, err}
	}()
	for {
		select {
		case <-stopChan:
			return waitIndex, nil
		case r := <-respChan:
			return r.waitIndex, r.err
		}
	}
}
