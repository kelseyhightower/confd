package metad

import (
	"container/ring"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kelseyhightower/confd/log"
)

type Connection struct {
	url        string
	transport  *http.Transport
	httpClient *http.Client
	waitIndex  uint64
	errTimes   uint32
}

func (c *Connection) testConnection() error {
	var err error
	maxTime := 15 * time.Second

	for i := 1 * time.Second; i < maxTime; i *= time.Duration(2) {
		if _, err = c.makeMetaDataRequest("/"); err != nil {
			time.Sleep(i)
		} else {
			return nil
		}
	}
	return err
}

func (c *Connection) makeMetaDataRequest(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", strings.Join([]string{c.url, path}, ""), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type Client struct {
	connections *ring.Ring
	current     *Connection
}

func NewMetadClient(backendNodes []string) (*Client, error) {
	if len(backendNodes) <= 0 {
		return nil,fmt.Errorf("no metad backend defined")
	}
	connections := ring.New(len(backendNodes))
	for _, backendNode := range backendNodes {
		url := "http://" + backendNode
		transport := &http.Transport{}
		connection := &Connection{
			url: url,
			httpClient: &http.Client{
				Transport: transport,
			},
			transport: transport,
		}
		connections.Value = connection
		connections = connections.Next()
	}

	client := &Client{
		connections: connections,
	}

	err := client.selectConnection()

	return client, err

}

func (c *Client) selectConnection() error {
	//random start
	if c.current == nil {
		rand.Seed(time.Now().Unix())
		r := rand.Intn(c.connections.Len())
		c.connections = c.connections.Move(r)
	}
	c.connections = c.connections.Next()
	conn := c.connections.Value.(*Connection)
	startConn := conn
	err := conn.testConnection()
	for err != nil {
		log.Error("Connection to [%s], error: [%s]", conn.url, err.Error())
		c.connections = c.connections.Next()
		conn = c.connections.Value.(*Connection)
		if conn == startConn {
			return errors.New("Fail to connect any backend.")
		}
		err = conn.testConnection()
	}
	if c.current != nil {
		atomic.StoreUint32(&c.current.errTimes, 0)
	}
	c.current = conn
	log.Info("Using Metad URL: " + c.current.url)
	return nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := map[string]string{}

	for _, key := range keys {
		body, err := c.current.makeMetaDataRequest(key)
		if err != nil {
			atomic.AddUint32(&c.current.errTimes, 1)
			return vars, err
		}

		var jsonResponse interface{}
		if err = json.Unmarshal(body, &jsonResponse); err != nil {
			return vars, err
		}

		if err = treeWalk(key, jsonResponse, vars); err != nil {
			return vars, err
		}
	}
	return vars, nil
}

func treeWalk(root string, val interface{}, vars map[string]string) error {
	switch val.(type) {
	case map[string]interface{}:
		for k := range val.(map[string]interface{}) {
			treeWalk(strings.Join([]string{root, k}, "/"), val.(map[string]interface{})[k], vars)
		}
	case []interface{}:
		for i, item := range val.([]interface{}) {
			idx := strconv.Itoa(i)
			if i, isMap := item.(map[string]interface{}); isMap {
				if name, exists := i["name"]; exists {
					idx = name.(string)
				}
			}

			treeWalk(strings.Join([]string{root, idx}, "/"), item, vars)
		}
	case bool:
		vars[root] = strconv.FormatBool(val.(bool))
	case string:
		vars[root] = val.(string)
	case float64:
		vars[root] = strconv.FormatFloat(val.(float64), 'f', -1, 64)
	case nil:
		vars[root] = "null"
	default:
		log.Error("Unknown type: " + reflect.TypeOf(val).Name())
	}
	return nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {

	if c.current.errTimes >= 3 {
		c.selectConnection()
	}

	conn := c.current

	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		conn.waitIndex = 1
		return conn.waitIndex, nil
	}
	// when switch to anther server, so set waitIndex 0, and let server response current version.
	if conn.waitIndex == 0 {
		waitIndex = 0
	}

	done := make(chan struct{})
	defer close(done)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?wait=true&prev_version=%d", conn.url, prefix, waitIndex), nil)
	if err != nil {
		return conn.waitIndex, err
	}
	req.Header.Set("Accept", "application/json")

	go func() {
		select {
		case <-stopChan:
			conn.transport.CancelRequest(req)
		case <-done:
			return
		}
	}()

	// just ignore resp, notify confd to reload metadata from metad
	resp, err := conn.httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		atomic.AddUint32(&conn.errTimes, 1)
		return conn.waitIndex, err
	}
	if resp.StatusCode != 200 {
		return conn.waitIndex, errors.New(fmt.Sprintf("metad response status [%v], requestID: [%s]", resp.StatusCode, resp.Header.Get("X-Metad-RequestID")))
	}
	versionStr := resp.Header.Get("X-Metad-Version")
	if versionStr != "" {
		v, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			log.Error("Parse X-Metad-Version %s error:%s", versionStr, err.Error())
		}
		conn.waitIndex = v
	} else {
		log.Warning("Metad response miss X-Metad-Version header.")
		conn.waitIndex = conn.waitIndex + 1
	}
	return conn.waitIndex, nil

}
