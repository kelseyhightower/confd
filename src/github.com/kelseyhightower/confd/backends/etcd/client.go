package etcd

import (
	"errors"
	"strings"
	"time"

	goetcd "github.com/coreos/go-etcd/etcd"
)

// Client is a wrapper around the etcd client
type Client struct {
	client *goetcd.Client
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewEtcdClient(machines []string, cert, key string, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	var c *goetcd.Client
	var err error
	if cert != "" && key != "" {
		c, err = goetcd.NewTLSClient(machines, cert, key, caCert)
		if err != nil {
			return &Client{c}, err
		}
	} else {
		c = goetcd.NewClient(machines)
	}
	// Configure BasicAuth if enabled
	if basicAuth {
		c.SetCredentials(username, password)
	}

	// Configure the DialTimeout, since 1 second is often too short
	c.SetDialTimeout(time.Duration(3) * time.Second)
	success := c.SetCluster(machines)
	if !success {
		return &Client{c}, errors.New("cannot connect to etcd cluster: " + strings.Join(machines, ","))
	}
	return &Client{c}, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string, token string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		resp, err := c.client.Get(key, true, true)
		if err != nil {
			return vars, err
		}
		err = nodeWalk(resp.Node, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node *goetcd.Node, vars map[string]string) error {
	if node != nil {
		key := node.Key
		if !node.Dir {
			vars[key] = node.Value
		} else {
			for _, node := range node.Nodes {
				nodeWalk(node, vars)
			}
		}
	}
	return nil
}

func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	if waitIndex == 0 {
		resp, err := c.client.Get(prefix, false, true)
		if err != nil {
			return 0, err
		}
		return resp.EtcdIndex, nil
	}
	resp, err := c.client.Watch(prefix, waitIndex+1, true, nil, stopChan)
	if err != nil {
		switch e := err.(type) {
		case *goetcd.EtcdError:
			if e.ErrorCode == 401 {
				return 0, nil
			}
		}
		return waitIndex, err
	}
	return resp.Node.ModifiedIndex, err
}
