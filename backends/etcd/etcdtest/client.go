// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package etcdtest

import (
	"github.com/coreos/go-etcd/etcd"
)

// Client represents a fake etcd client. Used for testing.
type Client struct {
	Responses map[string]*etcd.Response
}

// Get mimics the etcd.Client.Get() method.
func (c *Client) Get(key string, sort, recurse bool) (*etcd.Response, error) {
	return c.Responses[key], nil
}

// Watch mimics the etcd.Client.Watch() method
func (c *Client) Watch(key string, waitIndex uint64, recurse bool, receiver chan *etcd.Response, stop chan bool) (*etcd.Response, error) {
	receiver <- c.Responses[key]
	close(receiver)
	return nil, nil
}

// AddResponses adds or updates the Client.Responses map.
func (c *Client) AddResponse(key string, response *etcd.Response) {
	c.Responses[key] = response
}

// NewClient returns a fake etcd client.
func NewClient() *Client {
	responses := make(map[string]*etcd.Response)
	return &Client{responses}
}
