package vaultpki

import (
	"encoding/json"
	"errors"
	"path"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	vaultbackend "github.com/haad/confd/backends/vault"
	"github.com/haad/confd/log"
)

// Client Embed from vault client into vault pki client
type Client struct {
	*vaultapi.Client
}

// NewClient "inherit" the new method from the vault client
func NewClient(address, authType string, params map[string]string) (*Client, error) {
	if authType == "" {
		return nil, errors.New("you have to set the auth type when using the vault backend")
	}
	log.Info("Vault authentication backend set to %s", authType)
	conf, err := vaultbackend.GetConfig(address, params["cert"], params["key"], params["caCert"])

	if err != nil {
		return nil, err
	}

	c, err := vaultapi.NewClient(conf)
	if err != nil {
		return nil, err
	}

	if err := vaultbackend.Authenticate(c, authType, params); err != nil {
		return nil, err
	}

	return &Client{c}, err
}

// request this is the struct that will be sent to vault to issue a cert
type request struct {
	mountPath  string
	role       string
	commonName string
}

// parseCommonName this function parses the key return a request struct
func parseCommonName(key string) *request {
	// The key must be of the format /mountpath/issue/rolename/commonname
	if strings.Contains(key, "issue") {
		splitKeyList := strings.Split(key, "issue")
		splitRoleList := strings.Split(splitKeyList[1], "/")
		log.Debug("getCommonName path: %s, role: %s, commonName: %s", splitKeyList[0], splitRoleList[1], splitRoleList[2])
		k := request{splitKeyList[0], splitRoleList[1], splitRoleList[2]}
		return &k
	}

	return &request{}

}

func issueCert(c *Client, r *request) (map[string]string, error) {
	log.Debug("issueCert path: %s, role: %s, commonName: %s", r.mountPath, r.role, r.commonName)
	writePath := r.mountPath + "issue" + "/my-role"

	payload := map[string]interface{}{
		"common_name": r.commonName,
	}
	resp, err := c.Logical().Write(writePath, payload)
	vars := make(map[string]string)

	if err != nil {
		log.Debug("there was an error issuing a cert for role %s", r.role)
		return nil, err
	}

	// save the json encoded response
	// and flatten it to allow usage of gets & getvs
	js, _ := json.Marshal(resp.Data)
	vars[writePath] = string(js)
	vaultbackend.Flatten(writePath+"/"+r.commonName, resp.Data, vars)

	return vars, nil
}

func walkTree(c *Client, key string, branches map[string]bool) error {
	log.Debug("listing %s from vault", key)

	// strip trailing slash as long as it's not the only character
	if last := len(key) - 1; last > 0 && key[last] == '/' {
		key = key[:last]
	}
	if branches[key] {
		// already processed this branch
		return nil
	}
	branches[key] = true

	resp, err := c.Logical().List(key)

	if err != nil {
		log.Debug("there was an error extracting %s", key)
		return err
	}
	if resp == nil || resp.Data == nil || resp.Data["keys"] == nil {
		return nil
	}

	switch resp.Data["keys"].(type) {
	case []interface{}:
		// expected
	default:
		log.Warning("key list type of '%s' is not supported (%T)", key, resp.Data["keys"])
		return nil
	}

	keyList := resp.Data["keys"].([]interface{})
	for _, innerKey := range keyList {
		switch innerKey.(type) {

		case string:
			innerKey = path.Join(key, "/", innerKey.(string))
			walkTree(c, innerKey.(string), branches)

		default: // we don't know how to handle other data types
			log.Warning("type of '%s' is not supported (%T)", key, keyList)
		}
	}
	return nil
}

func (c *Client) getKvVault(keys []string) (map[string]string, error) {
	branches := make(map[string]bool)
	for _, key := range keys {
		walkTree(c, key, branches)
	}
	vars := make(map[string]string)
	for key := range branches {
		log.Debug("getting %s from vault", key)
		resp, err := c.Logical().Read(key)

		if err != nil {
			log.Debug("there was an error extracting %s", key)
			return nil, err
		}
		if resp == nil || resp.Data == nil {
			continue
		}

		// if the key has only one string value
		// treat it as a string and not a map of values
		if val, ok := vaultbackend.IsKV(resp.Data); ok {
			vars[key] = val
		} else {
			// save the json encoded response
			// and Flatten it to allow usage of gets & getvs
			js, _ := json.Marshal(resp.Data)
			vars[key] = string(js)
			vaultbackend.Flatten(key, resp.Data, vars)
		}
	}
	return vars, nil
}

// GetValues to be exported
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	var err error
	for _, key := range keys {
		log.Debug("key: %+v", key)
		k := parseCommonName(key)
		if (request{}) == *k {
			vars, err = c.getKvVault(keys)
			break
		}
		vars, err = issueCert(c, k)
	}

	return vars, err
}

// WatchPrefix - not implemented at the moment
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
