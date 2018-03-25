package etcdv3

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
	"github.com/kelseyhightower/confd/log"
)

// Client is a wrapper around the etcd client
type Client struct {
	cfg clientv3.Config
}

// NewEtcdClient returns an *etcdv3.Client with a connection to named machines.
func NewEtcdClient(machines []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	cfg := clientv3.Config{
		Endpoints:            machines,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
	}

	if basicAuth {
		cfg.Username = username
		cfg.Password = password
	}

	tlsEnabled := false
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if caCert != "" {
		certBytes, err := ioutil.ReadFile(caCert)
		if err != nil {
			return &Client{cfg}, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
		tlsEnabled = true
	}

	if cert != "" && key != "" {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return &Client{cfg}, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		tlsEnabled = true
	}

	if tlsEnabled {
		cfg.TLS = tlsConfig
	}

	return &Client{cfg}, nil
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

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	var err error

	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, err
	}

	client, err := clientv3.New(c.cfg)
	if err != nil {
		return 1, err
	}
	defer client.Close()

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

	rch := client.Watch(ctx, prefix, clientv3.WithPrefix())

	isChange := func(wresp clientv3.WatchResponse) int64 {
		for _, ev := range wresp.Events {
			log.Debug("Key updated %s", string(ev.Kv.Key))
			// Only return if we have a key prefix we care about.
			// This is not an exact match on the key so there is a chance
			// we will still pickup on false positives. The net win here
			// is reducing the scope of keys that can trigger updates.
			for _, k := range keys {
				if strings.HasPrefix(string(ev.Kv.Key), k) {
					return ev.Kv.Version
				}
			}
		}
		return -1
	}
	var index int64 = -1
	for {
		select {
		case <-time.After(time.Second * 5):
			if index == -1 {
				continue
			}
			return uint64(index), nil
		case wresp, ok := <-rch:
			if !ok {
				return 1, err
			}
			index = isChange(wresp)
		}
	}
	return 0, err
}
