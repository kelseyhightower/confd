package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/context"
)

// Client is a wrapper around the etcd client
type Client struct {
	client client.KeysAPI
}

// Configure configures etcd.Client with a connection to named machines.
func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}

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
		Endpoints:               strings.Split(config.Machines, ","),
		HeaderTimeoutPerRequest: time.Duration(3) * time.Second,
	}

	basicAuth, err := strconv.ParseBool(config.BasicAuth)
	if err != nil {
		return err
	}
	if basicAuth {
		cfg.Username = config.Username
		cfg.Password = config.Password
	}

	if config.CaCert != "" {
		certBytes, err := ioutil.ReadFile(config.CaCert)
		if err != nil {
			return err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
	}

	if config.Cert != "" && config.Key != "" {
		tlsCert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
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

func (c *Client) WatchPrefix(prefix string, keys []string, stream chan error) error {
	watcher := c.client.Watcher(prefix, &client.WatcherOptions{Recursive: true})
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			switch e := err.(type) {
			case *client.Error:
				if e.Code == http.StatusUnauthorized {
					return nil
				}
			}
			return err
		}

		// Only return if we have a key prefix we care about.
		// This is not an exact match on the key so there is a chance
		// we will still pickup on false positives. The net win here
		// is reducing the scope of keys that can trigger updates.
		for _, k := range keys {
			if strings.HasPrefix(resp.Node.Key, k) {
				stream <- nil
				break
			}
		}
	}
}
