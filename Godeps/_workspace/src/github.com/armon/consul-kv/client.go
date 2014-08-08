package consulkv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config is used to configure the creation of a client
type Config struct {
	// Address is the address of the Consul server
	Address string

	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string

	// HTTPClient is the client to use. Default will be
	// used if not provided.
	HTTPClient *http.Client

	// WaitTime limits how long a Watch will block. If not provided,
	// the agent default values will be used.
	WaitTime time.Duration
}

// Client provides a client to Consul for K/V data
type Client struct {
	config Config
}

// KVPair is used to represent a single K/V entry
type KVPair struct {
	Key         string
	CreateIndex uint64
	ModifyIndex uint64
	Flags       uint64
	Value       []byte
}

// KVPairs is a list of KVPair objects
type KVPairs []*KVPair

// KVMeta provides meta data about a query
type KVMeta struct {
	ModifyIndex uint64
}

// NewClient returns a new
func NewClient(config *Config) (*Client, error) {
	client := &Client{
		config: *config,
	}
	return client, nil
}

// DefaultConfig returns a default configuration for the client
func DefaultConfig() *Config {
	return &Config{
		Address:    "127.0.0.1:8500",
		HTTPClient: http.DefaultClient,
	}
}

// Get is used to lookup a single key
func (c *Client) Get(key string) (*KVMeta, *KVPair, error) {
	return selectOne(c.getRecurse(key, false, 0))
}

// List is used to lookup all keys with a prefix
func (c *Client) List(prefix string) (*KVMeta, KVPairs, error) {
	return c.getRecurse(prefix, true, 0)
}

// WatchGet is used to block and wait for a change on a key
func (c *Client) WatchGet(key string, modifyIndex uint64) (*KVMeta, *KVPair, error) {
	return selectOne(c.getRecurse(key, false, modifyIndex))
}

// WatchList is used to block and wait for a change on a prefix
func (c *Client) WatchList(prefix string, modifyIndex uint64) (*KVMeta, KVPairs, error) {
	return c.getRecurse(prefix, true, modifyIndex)
}

// deleteRecurse does a delete with a potential recurse
func (c *Client) getRecurse(key string, recurse bool, waitIndex uint64) (*KVMeta, KVPairs, error) {
	url := c.pathURL(key)
	query := url.Query()
	if recurse {
		query.Set("recurse", "1")
	}
	if waitIndex > 0 {
		query.Set("index", strconv.FormatUint(waitIndex, 10))
	}
	if waitIndex > 0 && c.config.WaitTime > 0 {
		waitMsec := fmt.Sprintf("%dms", c.config.WaitTime/time.Millisecond)
		query.Set("wait", waitMsec)
	}
	if len(query) > 0 {
		url.RawQuery = query.Encode()
	}
	req := http.Request{
		Method: "GET",
		URL:    url,
	}
	resp, err := c.config.HTTPClient.Do(&req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Decode the KVMeta
	meta := &KVMeta{}
	index, err := strconv.ParseUint(resp.Header.Get("X-Consul-Index"), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse X-Consul-Index: %v", err)
	}
	meta.ModifyIndex = index

	// Ensure status code is 404 or 200
	if resp.StatusCode == 404 {
		return meta, nil, nil
	} else if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	// Decode the response
	dec := json.NewDecoder(resp.Body)
	var out KVPairs
	if err := dec.Decode(&out); err != nil {
		return nil, nil, err
	}

	return meta, out, nil
}

// Put is used to set a value for a given key
func (c *Client) Put(key string, value []byte, flags uint64) error {
	_, err := c.putCAS(key, value, flags, 0, false)
	return err
}

// CAS is used for a Check-And-Set operation
func (c *Client) CAS(key string, value []byte, flags, index uint64) (bool, error) {
	return c.putCAS(key, value, flags, index, true)
}

// putCAS is used to do a PUT with optional CAS
func (c *Client) putCAS(key string, value []byte, flags, index uint64, cas bool) (bool, error) {
	url := c.pathURL(key)
	query := url.Query()
	if cas {
		query.Set("cas", strconv.FormatUint(index, 10))
	}
	query.Set("flags", strconv.FormatUint(flags, 10))
	url.RawQuery = query.Encode()
	req := http.Request{
		Method: "PUT",
		URL:    url,
		Body:   ioutil.NopCloser(bytes.NewReader(value)),
	}
	req.ContentLength = int64(len(value))
	resp, err := c.config.HTTPClient.Do(&req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return false, fmt.Errorf("failed to read response: %v", err)
	}
	res := strings.Contains(string(buf.Bytes()), "true")
	return res, nil
}

// Delete is used to delete a single key
func (c *Client) Delete(key string) error {
	return c.deleteRecurse(key, false)
}

// DeleteTree is used to delete all keys with a prefix
func (c *Client) DeleteTree(prefix string) error {
	return c.deleteRecurse(prefix, true)
}

// deleteRecurse does a delete with a potential recurse
func (c *Client) deleteRecurse(key string, recurse bool) error {
	url := c.pathURL(key)
	if recurse {
		query := url.Query()
		query.Set("recurse", "1")
		url.RawQuery = query.Encode()
	}
	req := http.Request{
		Method: "DELETE",
		URL:    url,
	}
	resp, err := c.config.HTTPClient.Do(&req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	return nil

}

// path is used to generate the HTTP path for a request
func (c *Client) pathURL(key string) *url.URL {
	url := &url.URL{
		Scheme: "http",
		Host:   c.config.Address,
		Path:   "/v1/kv/" + strings.TrimPrefix(key, "/"),
	}
	if c.config.Datacenter != "" {
		query := url.Query()
		query.Set("dc", c.config.Datacenter)
		url.RawQuery = query.Encode()
	}
	return url
}

// selectOne is used to grab only the first KVPair in a list
func selectOne(meta *KVMeta, pairs KVPairs, err error) (*KVMeta, *KVPair, error) {
	var pair *KVPair
	if len(pairs) > 0 {
		pair = pairs[0]
	}
	return meta, pair, err
}
