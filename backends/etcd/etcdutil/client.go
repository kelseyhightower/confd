// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package etcdutil

import (
	"errors"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// Client is a wrapper around the etcd client
type Client struct {
	client EtcdClient
}

type EtcdClient interface {
	Get(key string, sort, recurse bool) (*etcd.Response, error)
	Watch(key string, waitIndex uint64, recurse bool, receiver chan *etcd.Response, stop chan bool) (*etcd.Response, error)
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewEtcdClient(machines []string, cert, key string, caCert string) (*Client, error) {
	var c *etcd.Client
	var err error
	if cert != "" && key != "" {
		c, err = etcd.NewTLSClient(machines, cert, key, caCert)
		if err != nil {
			return &Client{c}, err
		}
	} else {
		c = etcd.NewClient(machines)
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
// Etcd paths (keys) are translated into names more suitable for use in
// templates. For example if prefix were set to '/production' and one of the
// keys were '/nginx/port'; the prefixed '/production/nginx/port' key would
// be queried for. If the value for the prefixed key where 80, the returned map
// would contain the entry vars["nginx_port"] = "80".
func (c *Client) GetValues(keys []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
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

// WatchValues watches etcd for updates to key prefixed by prefix.
// Etcd paths (keys) are translated into names more suitable for use in
// templates. For example if prefix were set to '/production' and one of the
// keys were '/nginx/port'; the prefixed '/production/nginx/port' key would
// be queried for. If the value for the prefixed key where 80, the returned map
// would contain the entry vars["nginx_port"] = "80".
func (c *Client) WatchValues(keys []string, varChan chan map[string]interface{}) error {
	receivers := map[string]chan *etcd.Response{}
	stop := make(chan bool)

	// close the channels when we're done with them
	defer close(varChan)
	defer close(stop)

	for _, key := range keys {
		// if the key already is in the map, ignore
		if _, ok := receivers[key]; !ok {
			// create channel for each key
			keyChan := make(chan *etcd.Response)
			receivers[key] = keyChan
			go c.client.Watch(key, 0, true, keyChan, stop)
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
				err := nodeWalk(varUpdate.Node, varUpdates)
				if err != nil {
					// stop all key watches
					stop <- true
					return err
				}
				varChan <- varUpdates

			}
		}
		if len(receivers) == 0 {
			return nil
		}
	}
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node *etcd.Node, vars map[string]interface{}) error {
	if node != nil {
		key := node.Key
		if !node.Dir {
			vars[key] = node.Value
		} else {
			vars[key] = node.Nodes
			for _, node := range node.Nodes {
				nodeWalk(node, vars)
			}
		}
	}
	return nil
}
