package json

import (
	"encoding/json"
	"io/ioutil"
	"strings"
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

	err = json.Unmarshal(fileContents, &allJsonVars)
	if err != nil {
		return &Client{allJsonVars}, err
	}

	return &Client{allJsonVars}, nil
}

// GetValues queries the json for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {

	vars := make(map[string]string)

	for _, key := range keys {
		jsonWalk(c.data, key, vars)
	}

	return vars, nil
}

// iterate over a json interface searching for key and add key/value to vars
func jsonWalk(json map[string]interface{}, key string, vars map[string]string) {
	keyPath := strings.Split(strings.TrimPrefix(key, "/"), "/")

	for index, component := range keyPath {
		if val, ok := json[component]; ok {
			// last component of key path
			if index == len(keyPath)-1 {
				if str, ok := val.(string); ok {
					// we have matched the key
					vars[key] = str
				} else {
					// we've hit the end of the key to match
					// but still have more object to traverse
					jsonAddAll(
						vars,
						strings.TrimPrefix(key, "/"),
						val.(map[string]interface{}))
				}
			} else {
				json = val.(map[string]interface{})
			}
		}
	}
}

// iterate over a json interface adding all data and adding to vars
func jsonAddAll(vars map[string]string, keyPrefix string, json map[string]interface{}) {

	for objkey, objvalue := range json {

		if str, ok := objvalue.(string); ok {
			// stop at strings and add a new entry to vars
			vars[keyPrefix+"/"+objkey] = str
		} else {
			// recurse
			jsonAddAll(
				vars,
				keyPrefix+"/"+objkey,
				objvalue.(map[string]interface{}))
		}
	}
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
