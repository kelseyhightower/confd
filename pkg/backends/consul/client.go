package consul

import (
	"path"
	"strings"

	"github.com/hashicorp/consul/api"
)

// Client provides a wrapper around the consulkv client
type ConsulClient struct {
	client *api.KV
}

// NewConsulClient returns a new client to Consul for the given address
func New(nodes []string, scheme, cert, key, caCert string, basicAuth bool, username string, password string) (*ConsulClient, error) {
	conf := api.DefaultConfig()

	conf.Scheme = scheme

	if len(nodes) > 0 {
		conf.Address = nodes[0]
	}

	if basicAuth {
		conf.HttpAuth = &api.HttpBasicAuth{
			Username: username,
			Password: password,
		}
	}

	if cert != "" && key != "" {
		conf.TLSConfig.CertFile = cert
		conf.TLSConfig.KeyFile = key
	}
	if caCert != "" {
		conf.TLSConfig.CAFile = caCert
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{client.KV()}, nil
}

// GetValues queries Consul for keys
func (c *ConsulClient) GetValues(keys []string) (map[string]string, error) {
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

func (c *ConsulClient) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	respChan := make(chan watchResponse)
	go func() {
		opts := api.QueryOptions{
			WaitIndex: waitIndex,
		}
		_, meta, err := c.client.List(prefix, &opts)
		if err != nil {
			respChan <- watchResponse{waitIndex, err}
			return
		}
		respChan <- watchResponse{meta.LastIndex, err}
	}()

	select {
	case <-stopChan:
		return waitIndex, nil
	case r := <-respChan:
		return r.waitIndex, r.err
	}
}
