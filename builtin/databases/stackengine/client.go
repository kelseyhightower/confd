package stackengine

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// Client is an empty wrapper around the StackEngine client
type Client struct {
	client    *http.Client
	token     string
	base      string
	transport *http.Transport
}

// Configure returns a client object with connection information.
func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}

	host := "127.0.0.1:8443"
	if len(strings.Split(config.Nodes, ",")) > 0 {
		host = strings.Split(config.Nodes, ",")[0]
	}
	c.base = config.Scheme + "://" + host
	c.token = config.AuthToken

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
	tlsConfig.InsecureSkipVerify = true
	c.transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	c.client = &http.Client{Transport: c.transport}
	return nil
}

// KVPair is used to represent a single K/V entry
type KVPair struct {
	Key         string
	CreateIndex uint64
	ModifyIndex uint64
	LockIndex   uint64
	Flags       uint64
	Value       []byte
	Session     string
}

// GetValues queries StackEngine for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	var pairs []KVPair

	for _, key := range keys {
		key := strings.TrimPrefix(key, "/")

		uri := c.base + "/v1/kv/" + key + "?recurse"
		req, err := http.NewRequest("GET", uri, nil)

		bearer := "Bearer " + c.token

		req.Header.Add("Authorization", bearer)
		parseFormErr := req.ParseForm()
		if parseFormErr != nil {
			fmt.Println(parseFormErr)
		}

		// Fetch Request
		resp, err := c.client.Do(req)

		if err != nil {
			fmt.Println("Failure : ", err)
			return nil, err
		}

		// Read Response Body
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading http response: ", err)
		}

		err = json.Unmarshal(respBody, &pairs)

		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[path.Join("/", p.Key)] = string(p.Value)
		}
	}

	return vars, nil
}

type watchResponse struct {
	waitIndex uint64
	err       error
}

func (c *Client) WatchPrefix(prefix string, keys []string, results chan string) error {
	return nil
}
