package vault

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/confd/confd"
	"github.com/mitchellh/mapstructure"
)

// Client is a wrapper around the vault client
type Client struct {
	client *vaultapi.Client
}

// Database returns a new client to Vault
func Database() confd.Database {
	return &Client{}
}

// Configure configures an *vault.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}

	if config.AuthType == "" {
		return errors.New("you have to set the auth type when using the vault backend")
	}
	log.Printf("[INFO] Vault authentication backend set to %s", config.AuthType)

	conf, err := getConfig(config.Address, config.Cert, config.Key, config.CaCert)
	if err != nil {
		return err
	}

	c.client, err = vaultapi.NewClient(conf)
	if err != nil {
		return err
	}

	if err = authenticate(c.client, config); err != nil {
		return err
	}
	return nil
}

// authenticate with the remote client
func authenticate(c *vaultapi.Client, config Config) (err error) {
	var secret *vaultapi.Secret

	switch config.AuthType {
	case "app-id":
		if config.AppId == "" {
			return errors.New("app-id is missing from configuration")
		} else if config.UserId == "" {
			return errors.New("user-id is missing from configuration")
		}
		secret, err = c.Logical().Write("/auth/app-id/login", map[string]interface{}{
			"app_id":  config.AppId,
			"user_id": config.UserId,
		})
	case "github":
		if config.Token == "" {
			return errors.New("token is missing from configuration")
		}
		secret, err = c.Logical().Write("/auth/github/login", map[string]interface{}{
			"token": config.Token,
		})
	case "token":
		if config.Token == "" {
			return errors.New("token is missing from configuration")
		}
		c.SetToken(config.Token)
		secret, err = c.Logical().Read("/auth/token/lookup-self")
	case "userpass":
		if config.Username == "" {
			return errors.New("username is missing from configuration")
		} else if config.Password == "" {
			return errors.New("password is missing from configuration")
		}
		secret, err = c.Logical().Write(fmt.Sprintf("/auth/userpass/login/%s", config.Username), map[string]interface{}{
			"password": config.Password,
		})
	}

	if err != nil {
		return err
	}

	// if the token has already been set
	if c.Token() != "" {
		return nil
	}

	log.Printf("[DEBUG] client authenticated with auth backend: %s", config.AuthType)
	// the default place for a token is in the auth section
	// otherwise, the backend will set the token itself
	c.SetToken(secret.Auth.ClientToken)
	return nil
}

func getConfig(address, cert, key, caCert string) (*vaultapi.Config, error) {
	conf := vaultapi.DefaultConfig()
	conf.Address = address

	tlsConfig := &tls.Config{}
	if cert != "" && key != "" {
		clientCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		tlsConfig.BuildNameToCertificate()
	}

	if caCert != "" {
		ca, err := ioutil.ReadFile(caCert)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(ca)
		tlsConfig.RootCAs = caCertPool
	}

	conf.HttpClient.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return conf, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		log.Printf("[DEBUG] getting %s from vault", key)
		resp, err := c.client.Logical().Read(key)

		if err != nil {
			log.Printf("[DEBUG] there was an error extracting %s", key)
			return nil, err
		}
		if resp == nil || resp.Data == nil {
			continue
		}

		// if the key has only one string value
		// treat it as a string and not a map of values
		if val, ok := isKV(resp.Data); ok {
			vars[key] = val
		} else {
			// save the json encoded response
			// and flatten it to allow usage of gets & getvs
			js, _ := json.Marshal(resp.Data)
			vars[key] = string(js)
			flatten(key, resp.Data, vars)
		}
	}
	return vars, nil
}

// isKV checks if a given map has only one key of type string
// if so, returns the value of that key
func isKV(data map[string]interface{}) (string, bool) {
	if len(data) == 1 {
		if value, ok := data["value"]; ok {
			if text, ok := value.(string); ok {
				return text, true
			}
		}
	}
	return "", false
}

// recursively walks on all the values of a specific key and set them in the variables map
func flatten(key string, value interface{}, vars map[string]string) {
	switch value.(type) {
	case string:
		log.Printf("[DEBUG] setting key %s to: %s", key, value)
		vars[key] = value.(string)
	case map[string]interface{}:
		inner := value.(map[string]interface{})
		for innerKey, innerValue := range inner {
			innerKey = path.Join(key, "/", innerKey)
			flatten(innerKey, innerValue, vars)
		}
	default: // we don't know how to handle non string or maps of strings
		log.Printf("[WARNING] type of '%s' is not supported (%T)", key, value)
	}
}

// WatchPrefix - not implemented at the moment
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64) (uint64, error) {
	return 0, nil
}
