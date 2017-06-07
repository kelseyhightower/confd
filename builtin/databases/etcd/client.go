package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/kelseyhightower/confd/confd"
	"golang.org/x/net/context"
)

// Client is a wrapper around the etcd client
type Client struct {
	client client.KeysAPI
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
func Database() confd.Database {
	return &Client{}
}

// Configure configures etcd.Client with a connection to named machines.
func (c *Client) Configure(config map[string]interface{}) (err error) {
	var transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	cfg := client.Config{
		Endpoints:               config["machines"].([]string),
		HeaderTimeoutPerRequest: time.Duration(3) * time.Second,
	}

	if config["basicAuth"].(bool) {
		cfg.Username = config["username"].(string)
		cfg.Password = config["password"].(string)
	}

	if config["caCert"].(string) != "" {
		certBytes, err := ioutil.ReadFile(config["caCert"].(string))
		if err != nil {
			return err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
	}

	if config["cert"].(string) != "" && config["key"].(string) != "" {
		tlsCert, err := tls.LoadX509KeyPair(config["cert"].(string), config["key"].(string))
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
	}

	transport.TLSClientConfig = tlsConfig
	cfg.Transport = transport

	etcdClient, err := client.New(cfg)
	if err != nil {
		return err
	}
	c.client = client.NewKeysAPI(etcdClient)
	return err
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		resp, err := c.client.Get(context.Background(), key, &client.GetOptions{
			Recursive: true,
			Sort:      true,
			Quorum:    true,
		})
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
func nodeWalk(node *client.Node, vars map[string]string) error {
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

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, nil
	}

	for {
		// Setting AfterIndex to 0 (default) means that the Watcher
		// should start watching for events starting at the current
		// index, whatever that may be.
		watcher := c.client.Watcher(prefix, &client.WatcherOptions{AfterIndex: uint64(0), Recursive: true})
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

		resp, err := watcher.Next(ctx)
		if err != nil {
			switch e := err.(type) {
			case *client.Error:
				if e.Code == 401 {
					return 0, nil
				}
			}
			return waitIndex, err
		}

		// Only return if we have a key prefix we care about.
		// This is not an exact match on the key so there is a chance
		// we will still pickup on false positives. The net win here
		// is reducing the scope of keys that can trigger updates.
		for _, k := range keys {
			if strings.HasPrefix(resp.Node.Key, k) {
				return resp.Node.ModifiedIndex, err
			}
		}
	}
}