package secretsmanager

//see https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/

// Secrets Manager does not have the equivelent of variables defined by paths eg /varroot/var1, /varoot/var2.
// Consequently we have to parse the secrets and look for '/' to emulate that functionality

import (
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/kelseyhightower/confd/log"
)

const delim = "/"

type Client struct {
	client *secretsmanager.SecretsManager
}

type SecretString struct {
	Name   string
	Secret string
}

var re = regexp.MustCompile("/[0-9A-Za-z_+=,.@-]+")

func New() (*Client, error) {
	// Create a session to share configuration, and load external configuration.
	sess := session.New()

	// Fail early, if no credentials can be found
	_, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	var c *aws.Config
	if os.Getenv("SECRETSMANAGER_LOCAL") != "" {
		log.Debug("SECRETSMANAGER_LOCAL is set")
		endpoint := "http://localhost:8002"
		c = &aws.Config{
			Endpoint: &endpoint,
		}
	}
	// Create the service's client with the session.
	svc := secretsmanager.New(sess, c)
	return &Client{svc}, nil
}

// GetValues retrieves the values for the given keys from AWS Secrets Manager
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	allkeys := make([]string, 0)
	knownkeys, err := c.buildNestedSecretsMap(keys)
	if err != nil {
		return vars, err
	}
	for _, key := range keys {
		log.Debug("Processing key=%s", key)
		if rootKey := re.FindString(key); len(rootKey) > 0 {
			allkeys = append(allkeys, knownkeys[rootKey]...)
			delete(knownkeys, rootKey)
		} else {
			allkeys = append(allkeys, key)
		}
	}
	for _, element := range allkeys {
		resp, err := c.getSecretValue(element)
		if err != nil {
			return vars, err
		}
		vars[resp.Name] = resp.Secret
	}
	return vars, nil
}

// buildNestedSecretsMap build a list of nested keys by calling describe keys
// and looking for keys of the format /x/*
func (c *Client) buildNestedSecretsMap(keys []string) (map[string][]string, error) {
	secrets := make(map[string][]string)
	param := &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int64(100),
	}
	resp, err := c.client.ListSecrets(param)
	if err != nil {
		return secrets, err
	}

	for _, element := range resp.SecretList {
		if rootKey := re.FindString(*element.Name); len(rootKey) > 0 {
			if secrets[rootKey] == nil {
				// create new slice with name
				secrets[rootKey] = []string{*element.Name}
			} else {
				//append to slice
				secrets[rootKey] = append(secrets[rootKey], *element.Name)
			}
		}
	}
	return secrets, err
}

// Retreive value from AWS
func (c *Client) getSecretValue(name string) (SecretString, error) {
	params := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"),
	}

	resp, err := c.client.GetSecretValue(params)
	if err != nil {
		return SecretString{}, err
	}
	secret := SecretString{
		Name:   *resp.Name,
		Secret: *resp.SecretString,
	}
	return secret, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
