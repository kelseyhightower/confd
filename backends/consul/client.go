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

// WatchValues watches Consul for key changes and directs them to the channel
func (c *Client) WatchValues(keys []string, vars chan map[string]interface{}) error {
	receivers := make(map[string]chan map[string]interface{})

	// close the channels when we're done with them
	defer close(vars)

	for _, key := range keys {
		// if the key already is in the map, ignore
		if _, ok := receivers[key]; !ok {
			// create channel for each key
			keyChan := make(chan map[string]interface{})
			receivers[key] = keyChan
			go c.watchKey(key, keyChan)
		}
	}

	varUpdates := make(map[string]interface{})

	for {
		for key := range receivers {
			select {
			case varUpdate, ok := <-receivers[key]:
				if !ok {
					delete(receivers, key)
					continue
				}
				for key, value := range varUpdate {
					varUpdates[key] = value
				}
				vars <- varUpdates
			}
		}
		if len(receivers) == 0 {
			return nil
		}
	}
}

func (c *Client) watchKey(key string, varChan chan map[string]interface{}) error {
	vars := make(map[string]interface{})
	defer close(varChan)
	for {
		_, pairs, err := c.client.WatchList(key, 0)
		if err != nil {
			return err
		}
		for _, p := range pairs {
			vars[p.Key] = string(p.Value)
		}
		varChan <- vars
	}
}
