package redis

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/util"
	"github.com/garyburd/redigo/redis"
)

// Client is a wrapper around the redis client
type Client struct {
	client redis.Conn
}

// NewRedisClient returns an *redis.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewRedisClient(rc *config.RedisBackendConfig) (*Client, error) {
	machines := util.GetBackendNodesFromSRVOrElse(rc.Type(), rc.Srv, func()[]string { return rc.Nodes })
	var err error
	for _, address := range machines {
		var conn redis.Conn
		network := "tcp"
		if _, err = os.Stat(address); err == nil {
			network = "unix"
		}
		conn, err = redis.DialTimeout(network, address, time.Second, time.Second, time.Second)
		if err != nil {
			continue
		}
		return &Client{conn}, nil
	}
	return nil, err
}

// GetValues queries redis for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		key = strings.Replace(key, "/*", "", -1)
		value, err := redis.String(c.client.Do("GET", key))
		if err == nil {
			vars[key] = value
			continue
		}

		if err != redis.ErrNil {
			return vars, err
		}

		if key == "/" {
			key = "/*"
		} else {
			key = fmt.Sprintf("%s/*", key)
		}

		idx := 0
		for {
			values, err := redis.Values(c.client.Do("SCAN", idx, "MATCH", key, "COUNT", "1000"))
			if err != nil && err != redis.ErrNil {
				return vars, err
			}
			idx, _ = redis.Int(values[0], nil)
			items, _ := redis.Strings(values[1], nil)
			for _, item := range items {
				var newKey string
				if newKey, err = redis.String(item, nil); err != nil {
					return vars, err
				}
				if value, err = redis.String(c.client.Do("GET", newKey)); err == nil {
					vars[newKey] = value
				}
			}
			if idx == 0 {
				break
			}
		}
	}
	return vars, nil
}

// WatchPrefix is not yet implemented.
func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
