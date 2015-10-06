package etcdcrypt

import (
	"errors"
	"strings"
	"time"
	"os"

	crypt "github.com/xordataexchange/crypt/encoding/secconf"
	goetcd "github.com/coreos/go-etcd/etcd"

	"github.com/kelseyhightower/confd/log"
)

// Client is a wrapper around the etcd client
type Client struct {
	client *goetcd.Client
}

var KEY_FILE string

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewEtcdClient(machines []string, cert, key string, caCert string, secKeyFile string) (*Client, error) {
	var c *goetcd.Client
	var err error
	var kr *os.File
	if cert != "" && key != "" {
		c, err = goetcd.NewTLSClient(machines, cert, key, caCert)
		if err != nil {
			return &Client{c}, err
		}
	} else {
		c = goetcd.NewClient(machines)
	}
	kr, err = os.Open(secKeyFile)
	if err != nil {
		return &Client{c}, errors.New("Cannot open secret key file: " + secKeyFile)
	}
        defer kr.Close()
	KEY_FILE = secKeyFile
	// Configure the DialTimeout, since 1 second is often too short
	c.SetDialTimeout(time.Duration(3) * time.Second)
	success := c.SetCluster(machines)
	if !success {
		return &Client{c}, errors.New("cannot connect to etcd cluster: " + strings.Join(machines, ","))
	}
	return &Client{c}, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		resp, err := c.client.Get(key, true, true)
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
func nodeWalk(node *goetcd.Node, vars map[string]string) error {
	kr, err := os.Open(KEY_FILE)
	if err != nil {
		log.Error("could not open secret keyfile " + KEY_FILE)
		return errors.New("could not open secret keyfile")
	}
	defer kr.Close()
	if node != nil {
		key := node.Key
		if !node.Dir {
			raw, err := crypt.Decode([]byte(node.Value), kr)
			if err != nil {
                                log.Error("crypt.Decode failed")
                                return err
			} else {
				vars[key] = strings.TrimRight(string(raw[:]), "\n")
			}
		} else {
			for _, node := range node.Nodes {
				nodeWalk(node, vars)
			}
		}
	}
	return nil
}

func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	if waitIndex == 0 {
		resp, err := c.client.Get(prefix, false, true)
		if err != nil {
			return 0, err
		}
		return resp.EtcdIndex, nil
	}
	resp, err := c.client.Watch(prefix, waitIndex+1, true, nil, stopChan)
	if err != nil {
		switch e := err.(type) {
		case *goetcd.EtcdError:
			if e.ErrorCode == 401 {
				return 0, nil
			}
		}
		return waitIndex, err
	}
	return resp.Node.ModifiedIndex, err
}
