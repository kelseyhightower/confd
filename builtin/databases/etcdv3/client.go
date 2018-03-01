package etcdv3

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
	"github.com/mitchellh/mapstructure"
)

// Client is a wrapper around the etcd client
type Client struct {
	cfg clientv3.Config
}

func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}

	c.cfg = clientv3.Config{
		Endpoints:            strings.Split(config.Machines, ","),
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
	}

	basicAuth, err := strconv.ParseBool(config.BasicAuth)
	if err != nil {
		return err
	}
	if basicAuth {
		c.cfg.Username = config.Username
		c.cfg.Password = config.Password
	}

	tlsEnabled := false
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
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
		tlsEnabled = true
	}

	if config.Cert != "" && config.Key != "" {
		tlsCert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		tlsEnabled = true
	}

	if tlsEnabled {
		c.cfg.TLS = tlsConfig
	}

	return nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)

	client, err := clientv3.New(c.cfg)
	if err != nil {
		return vars, err
	}
	defer client.Close()

	for _, key := range keys {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
		resp, err := client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
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

func (c *Client) WatchPrefix(prefix string, keys []string, results chan string) error {
	client, err := clientv3.New(c.cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	rch := client.Watch(context.Background(), prefix, clientv3.WithPrefix())

	for wresp := range rch {
		for _, ev := range wresp.Events {
			log.Printf("[DEBUG] Key updated %s", string(ev.Kv.Key))
			// Only return if we have a key prefix we care about.
			// This is not an exact match on the key so there is a chance
			// we will still pickup on false positives. The net win here
			// is reducing the scope of keys that can trigger updates.
			for _, k := range keys {
				if strings.HasPrefix(string(ev.Kv.Key), k) {
					results <- ""
					break
				}
			}
		}
	}
	return nil
}
