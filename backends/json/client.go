package json

import (
	"encoding/json"
	"io/ioutil"
)

// Client provides a shell for the json client
type Client struct {
	data map[string]interface{}
}

// NewJsonClient returns a new client
func NewJsonClient(filePath string) (*Client, error) {
	var allJsonVars map[string]interface{}

	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return &Client{allJsonVars}, err
	}

	err = json.Unmarshal(fileContents, allJsonVars)
	if err != nil {
		return &Client{allJsonVars}, err
	}

	return &Client{allJsonVars}, nil
}

// GetValues queries the json for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {

	vars := make(map[string]string)

	for _, key := range keys {
		vars[key] = c.data[key].(string)
	}

	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
