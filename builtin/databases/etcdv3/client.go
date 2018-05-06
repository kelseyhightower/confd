package etcdv3

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/kelseyhightower/confd/log"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/context"
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

	// Use all operations on the same revision
	var first_rev int64 = 0
	// Default ETCDv3 TXN limitation. Since it is configurable from v3.3,
	// maybe an option should be added (also set max-txn=0 can disable Txn?)
	maxTxnOps := 128
	getOps := make([]string, 0, maxTxnOps)
	doTxn := func(ops []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
		defer cancel()

		txnOps := make([]clientv3.Op, 0, maxTxnOps)

		for _, k := range ops {
			txnOps = append(txnOps, clientv3.OpGet(k,
				clientv3.WithPrefix(),
				clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
				clientv3.WithRev(first_rev)))
		}

		result, err := client.Txn(ctx).Then(txnOps...).Commit()
		if err != nil {
			return err
		}
		for i, r := range result.Responses {
			originKey := ops[i]
			// append a '/' if not already exists
			originKeyFixed := originKey
			if !strings.HasSuffix(originKeyFixed, "/") {
				originKeyFixed = originKey + "/"
			}
			for _, ev := range r.GetResponseRange().Kvs {
				k := string(ev.Key)
				if k == originKey || strings.HasPrefix(k, originKeyFixed) {
					vars[string(ev.Key)] = string(ev.Value)
				}
			}
		}
		if first_rev == 0 {
			// Save the revison of the first request
			first_rev = result.Header.GetRevision()
		}
		return nil
	}
	for _, key := range keys {
		getOps = append(getOps, key)
		if len(getOps) >= maxTxnOps {
			if err := doTxn(getOps); err != nil {
				return vars, err
			}
			getOps = getOps[:0]
		}
	}
	if len(getOps) > 0 {
		if err := doTxn(getOps); err != nil {
			return vars, err
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
			log.Debug("Key updated %s", string(ev.Kv.Key))
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
