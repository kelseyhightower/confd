package consul

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/kelseyhightower/confd/log"
	"github.com/mitchellh/mapstructure"
)

// Client provides a wrapper around the consulkv client
type Client struct {
	client *api.KV
}

func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}

	conf := api.DefaultConfig()

	conf.Scheme = config.Scheme

	if len(strings.Split(config.Nodes, ",")) > 0 {
		conf.Address = strings.Split(config.Nodes, ",")[0]
	}

	tlsConfig := &tls.Config{}
	if config.Cert != "" && config.Key != "" {
		clientCert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		tlsConfig.BuildNameToCertificate()
	}
	if config.CaCert != "" {
		ca, err := ioutil.ReadFile(config.CaCert)
		if err != nil {
			return err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(ca)
		tlsConfig.RootCAs = caCertPool
	}
	conf.Transport.TLSClientConfig = tlsConfig
	conf.HttpClient = &http.Client{
		Transport: conf.Transport,
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return err
	}
	c.client = client.KV()
	return nil
}

// GetValues queries Consul for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		key = strings.TrimPrefix(key, "/")
		pairs, _, err := c.client.List(key, nil)
		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[path.Join("/", p.Key)] = string(p.Value)
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, results chan string) error {
	var index uint64
	for {
		_, meta, err := c.client.List(prefix, &api.QueryOptions{WaitIndex: index})
		if err != nil {
			log.Error(err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		if meta.LastIndex != index {
			index = meta.LastIndex
			results <- ""
		}
	}
}
