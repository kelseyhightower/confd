package etcdv3

import (
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"golang.org/x/net/context"
)

// Client is a wrapper around the etcd client
type Client struct {
	client *clientv3.Client
}

// NewEtcdClientv3 returns an *etcd.Client with a connection to named machines.
func NewEtcdClientv3(machines []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	var c *clientv3.Client
	var err error

	cfg := clientv3.Config{
		Endpoints:   machines,
		DialTimeout: 30 * time.Second,
	}

	if basicAuth {
		cfg.Username = username
		cfg.Password = password
	}

	if cert != "" && key != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      cert,
			KeyFile:       key,
			TrustedCAFile: caCert,
		}

		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return &Client{c}, err
		}

		cfg.TLS = tlsConfig
	}

	c, err = clientv3.New(cfg)
	if err != nil {
		return &Client{c}, err
	}
	return &Client{c}, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		ctx, cancel := context.WithTimeout(context.Background(),
			3*time.Second)

		resp, err := c.client.Get(ctx, key,
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		cancel()
		if err != nil {
			return vars, err
		}

		for _, ev := range resp.Kvs {
			vars[string(ev.Key)] = string(ev.Value)
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, nil
	}

	for {
		// Setting AfterIndex to 0 (default) means that the Watcher
		// should start watching for events starting at the current
		// index, whatever that may be.

		ctx, cancel := context.WithCancel(context.Background())
		cancelRoutine := make(chan bool)
		defer close(cancelRoutine)

		go func() {
			select {
			case <-stopChan:
				cancel()
			case <-cancelRoutine:
				return
			}
		}()

		// Only return if we have a key prefix we care about.
		// This is not an exact match on the key so there is a chance
		// we will still pickup on false positives. The net win here
		// is reducing the scope of keys that can trigger updates.
		rch := c.client.Watch(ctx, prefix, clientv3.WithPrefix())
		for wresp := range rch {
			for _, ev := range wresp.Events {
				for _, k := range keys {
					if strings.HasPrefix(string(ev.Kv.Key), k) {
						return uint64(ev.Kv.Version), nil
					}
				}
			}
		}
	}
}
