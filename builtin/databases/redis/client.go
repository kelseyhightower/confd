package redis

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/kelseyhightower/confd/log"
	"github.com/mitchellh/mapstructure"
)

type watchResponse struct {
	waitIndex uint64
	err       error
}

// Client is a wrapper around the redis client
type Client struct {
	client    redis.Conn
	machines  []string
	password  string
	separator string
	psc       redis.PubSubConn
	pscChan   chan watchResponse
}

func (c *Client) Configure(configRaw map[string]string) error {
	var config Config

	err := mapstructure.Decode(configRaw, &config)
	if err != nil {
		return err
	}

	if config.Separator == "" {
		config.Separator = "/"
	}
	log.Debug("Redis Separator: %#v", config.Separator)
	c.machines = strings.Split(config.Machines, ",")
	c.password = config.Password
	c.separator = config.Separator
	c.pscChan = make(chan watchResponse)
	c.psc = redis.PubSubConn{Conn: nil}
	c.client, _, err = tryConnect(c.machines, c.password, true)
	return err
}

// Iterate through `machines`, trying to connect to each in turn.
// Returns the first successful connection or the last error encountered.
// Assumes that `machines` is non-empty.
func tryConnect(machines []string, password string, timeout bool) (redis.Conn, int, error) {
	var err error
	for _, address := range machines {
		var conn redis.Conn
		var db int

		idx := strings.Index(address, "/")
		if idx != -1 {
			// a database is provided
			db, err = strconv.Atoi(address[idx+1:])
			if err == nil {
				address = address[:idx]
			}
		}

		network := "tcp"
		if _, err = os.Stat(address); err == nil {
			network = "unix"
		}
		log.Debug("Trying to connect to redis node %s", address)

		var dialops []redis.DialOption
		if timeout {
			dialops = []redis.DialOption{
				redis.DialConnectTimeout(time.Second),
				redis.DialReadTimeout(time.Second),
				redis.DialWriteTimeout(time.Second),
				redis.DialDatabase(db),
			}
		} else {
			dialops = []redis.DialOption{
				redis.DialConnectTimeout(time.Second),
				redis.DialWriteTimeout(time.Second),
				redis.DialDatabase(db),
			}
		}

		if password != "" {
			dialops = append(dialops, redis.DialPassword(password))
		}

		conn, err = redis.Dial(network, address, dialops...)

		if err != nil {
			continue
		}
		return conn, db, nil
	}
	return nil, 0, err
}

// Retrieves a connected redis client from the client wrapper.
// Existing connections will be tested with a PING command before being returned. Tries to reconnect once if necessary.
// Returns the established redis connection or the error encountered.
func (c *Client) connectedClient() (redis.Conn, error) {
	if c.client != nil {
		log.Debug("Testing existing redis connection.")

		resp, err := c.client.Do("PING")
		if (err != nil && err == redis.ErrNil) || resp != "PONG" {
			log.Error("Existing redis connection no longer usable. "+
				"Will try to re-establish. Error: %s", err.Error())
			c.client = nil
		}
	}

	// Existing client could have been deleted by previous block
	if c.client == nil {
		var err error
		c.client, _, err = tryConnect(c.machines, c.password, true)
		if err != nil {
			return nil, err
		}
	}

	return c.client, nil
}

func (c *Client) transform(key string) string {
	if c.separator == "/" {
		return key
	}
	k := strings.TrimPrefix(key, "/")
	return strings.Replace(k, "/", c.separator, -1)
}

func (c *Client) clean(key string) string {
	k := key
	if !strings.HasPrefix(k, "/") {
		k = "/" + k
	}
	return strings.Replace(k, c.separator, "/", -1)
}

// GetValues queries redis for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	// Ensure we have a connected redis client
	rClient, err := c.connectedClient()
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	vars := make(map[string]string)
	for _, key := range keys {
		key = strings.Replace(key, "/*", "", -1)

		k := c.transform(key)
		t, err := redis.String(rClient.Do("TYPE", k))

		if err == nil && err != redis.ErrNil {

			if t == "string" {
				value, err := redis.String(rClient.Do("GET", k))
				if err == nil {
					vars[key] = value
					continue
				}
				if err != redis.ErrNil {
					return vars, err
				}
			} else if t == "hash" {
				idx := 0
				for {
					values, err := redis.Values(rClient.Do("HSCAN", k, idx, "MATCH", "*", "COUNT", "1000"))
					if err != nil && err != redis.ErrNil {
						return vars, err
					}
					idx, _ = redis.Int(values[0], nil)
					items, _ := redis.Strings(values[1], nil)
					for i := 0; i < len(items); i += 2 {
						var newKey, value string
						if newKey, err = redis.String(items[i], nil); err != nil {
							return vars, err
						}
						if value, err = redis.String(items[i+1], nil); err != nil {
							return vars, err
						}
						vars[c.clean(k+"/"+newKey)] = value
					}
					if idx == 0 {
						break
					}
				}
			} else {
				if key == "/" {
					k = "*"
				} else {
					k = fmt.Sprintf(c.transform("%s/*"), k)
				}

				idx := 0
				for {
					values, err := redis.Values(rClient.Do("SCAN", idx, "MATCH", k, "COUNT", "1000"))
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
						if value, err := redis.String(rClient.Do("GET", newKey)); err == nil {
							vars[c.clean(newKey)] = value
						}
					}
					if idx == 0 {
						break
					}
				}
			}
		} else {
			return vars, err
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, results chan string) error {
	rClient, db, err := tryConnect(c.machines, c.password, false)
	if err != nil {
		return err
	}

	psc := redis.PubSubConn{Conn: rClient}
	defer psc.Close()

	psc.PSubscribe("__keyspace@" + strconv.Itoa(db) + "__:" + c.transform(prefix) + "*")
	defer psc.PUnsubscribe()

	for {
		switch n := psc.Receive().(type) {
		case redis.PMessage:
			log.Error("Redis Message: %s %s", n.Channel, n.Data)
			data := string(n.Data)
			commands := [12]string{"del", "append", "rename_from", "rename_to", "expire", "set", "incrby", "incrbyfloat", "hset", "hincrby", "hincrbyfloat", "hdel"}
			for _, command := range commands {
				if command == data {
					results <- ""
					break
				}
			}
		case redis.Subscription:
			log.Debug("Redis Subscription: %s %s %d", n.Kind, n.Channel, n.Count)
			if n.Count == 0 {
				results <- ""
			}
		case error:
			log.Error("Redis error: %s", n.Error())
			time.Sleep(2 * time.Second)
		}
	}
}
