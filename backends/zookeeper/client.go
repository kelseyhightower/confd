package zookeeper

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/kelseyhightower/confd/log"
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
	waitIndex uint64
	err       error
}

func (c *Client) watchFolder(folder string, respChan chan watchResponse, cancelRoutine chan bool) {
	_, _, eventCh, err := c.client.ChildrenW(folder)
	if err != nil {
		respChan <- watchResponse{0, err}
	}
	for {
		select {
		case e := <-eventCh:
			if e.Type == zk.EventNodeChildrenChanged {
				respChan <- watchResponse{1, e.Err}
			}
		case <-cancelRoutine:
			// There is no way to stop ChildrenW so just quit
			return
		}
	}
}

func (c *Client) watchKey(key string, respChan chan watchResponse, cancelRoutine chan bool) {
	_, _, eventCh, err := c.client.GetW(key)
	if err != nil {
		respChan <- watchResponse{0, err}
	}
	for {
		select {
		case e := <-eventCh:
			if e.Type == zk.EventNodeDataChanged {
				respChan <- watchResponse{1, e.Err}
			}
		case <-cancelRoutine:
			// There is no way to stop GetW so just quit
			return
		}
	}
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, nil
	}

	// List the childrens first
	entries, err := c.GetValues([]string{prefix})
	if err != nil {
		return 0, err
	}

	respChan := make(chan watchResponse)
	cancelRoutine := make(chan bool)
	defer close(cancelRoutine)

	//watch prefix for new/deleted children
	watchMap := make(map[string]string)
	log.Debug("Watching: " + prefix)
	go c.watchFolder(prefix, respChan, cancelRoutine)
	//watch all subfolders for changes
	for k, _ := range entries {
		for dir := filepath.Dir(k); dir != prefix; dir = filepath.Dir(dir) {
			if _, ok := watchMap[dir]; !ok {
				watchMap[dir] = ""
				log.Debug("Watching: " + dir)
				go c.watchFolder(dir, respChan, cancelRoutine)
			}
		}
	}

	//watch all keys in prefix for changes
	for k, _ := range entries {
		log.Debug("Watching: " + k)
		go c.watchKey(k, respChan, cancelRoutine)
	}

	for {
		select {
		case <-stopChan:
			return waitIndex, nil
		case r := <-respChan:
			return r.waitIndex, r.err
		}
	}
}
