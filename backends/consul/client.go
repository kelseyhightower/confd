package consul

import (
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
		pairs, _, err := c.client.List(key, nil)
		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[p.Key] = string(p.Value)
		}
	}
	return vars, nil
}
