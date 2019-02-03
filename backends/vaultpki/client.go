package vaultpki

import (
	"errors"

	vaultclient "github.com/kelseyhightower/confd/backends/vault"
	"github.com/kelseyhightower/confd/log"
)

// Client is a wrapper around vault client
type Client struct {
	client *vaultapi.Client
}

func

// New Connect to the vault instance and return a *vault.Client connection
func New(address, authType string, params map[string]string) (*Client, error) {
	if authType == "" {
		return nil, errors.New("You have to set the auth type when using the vault backend")
	}
	log.Info("Vault authentication backend set to %s", authType)
	conf, err := getConfig

	return nil, nil
}

//GetValues  queries vault and gets the cert
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	return nil, nil
}

// WatchPrefix not yet implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
