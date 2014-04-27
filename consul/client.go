package consul

import (
	"github.com/armon/consul-kv"
)

// Client provides a wrapper around the consulkv client
type Client struct {
	client *consulkv.Client
}

// NewConsulClient returns a new client to Consul for the given address
func NewConsulClient(addr string) (*Client, error) {
	conf := consulkv.DefaultConfig()
	conf.Address = addr
	client, err := consulkv.NewClient(conf)
	if err != nil {
		return nil, err
	}
	c := &Client{
		client: client,
	}
	return c, nil
}

// GetValues queries Consul for keys
func (c *Client) GetValues(keys []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	for _, key := range keys {
		_, pairs, err := c.client.List(key)
		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[p.Key] = string(p.Value)
		}
	}
	return vars, nil
}
