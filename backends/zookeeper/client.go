package zookeeper

import (
	"strings"
	"time"

	zk "github.com/samuel/go-zookeeper/zk"
)

// Client provides a wrapper around the zookeeper client
type Client struct {
	client *zk.Conn
}

func NewZookeeperClient(machines []string) (*Client, error) {
	c, _, err := zk.Connect(machines, time.Second) //*10)
	if err != nil {
		panic(err)
	}
	return &Client{c}, nil
}

func nodeWalk(prefix string, c *Client, vars map[string]string) error {
	l, stat, err := c.client.Children(prefix)
	if err != nil {
		return err
	}

	if stat.NumChildren == 0 {
		b, _, err := c.client.Get(prefix)
		if err != nil {
			return err
		}
		vars[prefix] = string(b)

	} else {
		for _, key := range l {
			s := prefix + "/" + key
			_, stat, err := c.client.Exists(s)
			if err != nil {
				return err
			}
			if stat.NumChildren == 0 {
				b, _, err := c.client.Get(s)
				if err != nil {
					return err
				}
				vars[s] = string(b)
			} else {
				nodeWalk(s, c, vars)
			}
		}
	}
	return nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, v := range keys {
		v = strings.Replace(v, "/*", "", -1)
		_, _, err := c.client.Exists(v)
		if err != nil {
			return vars, err
		}
		if v == "/" {
			v = ""
		}
		err = nodeWalk(v, c, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

type watchResponse struct {
	event zk.Event
	err   error
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, nil
	}

	watchChan := make(chan watchResponse)
	cancelRoutine := make(chan bool)
	defer close(cancelRoutine)

	for _, v := range keys {
		go func(key string) {
			_, _, eventChan, err := c.client.GetW(key)
			if err != nil {
				watchChan <- watchResponse{zk.Event{}, err}
				return
			} else {
				select {
				case event := <-eventChan:
					watchChan <- watchResponse{event, nil}
				case <-cancelRoutine:
					return
				}
			}
		}(v)
	}

	select {
	case response := <-watchChan:
		if response.err != nil {
			return 0, response.err
		}
		if response.event.Type == zk.EventNodeDataChanged {
			return 1, nil
		}
	case <-stopChan:
		return 0, nil
	}

	return 0, nil
}
