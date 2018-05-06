package zookeeper

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/kelseyhightower/confd/log"
	"github.com/mitchellh/mapstructure"
	zk "github.com/samuel/go-zookeeper/zk"
)

// Client provides a wrapper around the zookeeper client
type Client struct {
	client *zk.Conn
}

func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}

	client, _, err := zk.Connect(strings.Split(config.Machines, ","), time.Second) //*10)
	c.client = client
	return err
}

func nodeWalk(prefix string, c *Client, vars map[string]string) error {
	var s string
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
			if prefix == "/" {
				s = "/" + key
			} else {
				s = prefix + "/" + key
			}
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
		err = nodeWalk(v, c, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

func (c *Client) watch(key string, respChan chan error, cancelRoutine chan bool) {
	_, _, keyEventCh, err := c.client.GetW(key)
	if err != nil {
		respChan <- err
	}
	_, _, childEventCh, err := c.client.ChildrenW(key)
	if err != nil {
		respChan <- err
	}

	for {
		select {
		case e := <-keyEventCh:
			if e.Type == zk.EventNodeDataChanged {
				respChan <- e.Err
			}
		case e := <-childEventCh:
			if e.Type == zk.EventNodeChildrenChanged {
				respChan <- e.Err
			}
		case <-cancelRoutine:
			log.Debug("Stop watching: %s", key)
			// There is no way to stop GetW/ChildrenW so just quit
			return
		}
	}
}

func (c *Client) WatchPrefix(prefix string, keys []string, results chan string) error {
	// List the childrens first
	entries, err := c.GetValues([]string{prefix})
	if err != nil {
		return err
	}

	respChan := make(chan error)
	cancelRoutine := make(chan bool)
	defer close(cancelRoutine)

	//watch all subfolders for changes
	watchMap := make(map[string]string)
	for k := range entries {
		for _, v := range keys {
			if strings.HasPrefix(k, v) {
				for dir := filepath.Dir(k); dir != "/"; dir = filepath.Dir(dir) {
					if _, ok := watchMap[dir]; !ok {
						watchMap[dir] = ""
						log.Debug("Watching: %s", dir)
						go c.watch(dir, respChan, cancelRoutine)
					}
				}
				break
			}
		}
	}

	//watch all keys in prefix for changes
	for k := range entries {
		for _, v := range keys {
			if strings.HasPrefix(k, v) {
				log.Debug("Watching: %s", k)
				go c.watch(k, respChan, cancelRoutine)
				break
			}
		}
	}

	for {
		err := <-respChan
		if err != nil {
			log.Error(err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		results <- ""
	}
}
