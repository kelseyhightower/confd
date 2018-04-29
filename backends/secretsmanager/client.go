package secretsmanager

//see https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/

// Secrets Manager does not have the equivelent of variables defined by paths eg /varroot/var1, /varoot/var2.
// Consequently we have to parse the secrets and look for '/' to emulate that functionality

import (
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/kelseyhightower/confd/log"
)

type Client struct {
	client *secretsmanager.SecretsManager
}

type SecretString struct {
	Name   string
	Secret string
}

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
	knownkeys, err := c.buildNestedSecretsSlice()
	if err != nil {
		return vars, err
	}
	sort.Strings(knownkeys)
	sort.Strings(keys)
	log.Debug("known keys =%v \n keys =%v", knownkeys, keys)

	ilen := len(keys)
	klen := len(knownkeys)
	i := 0
	// This works because both sets of keys are sorted
	// and we keep a reference of where we are up to
	for k := 0; k < klen && i < ilen; {
		found := false
		for k < klen && i < ilen && strings.HasPrefix(knownkeys[k], keys[i]) {
			found = true
			allkeys = append(allkeys, knownkeys[k])
			k++
		}
		if k >= klen || i >= ilen {
			break
		} else if found || keys[i] < knownkeys[k] {
			// increment the key
			i++
		} else {
			// go to the next known key
			k++
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

// buildNestedSecretsSlice build a slice of nested keys by calling describe keys
// and looking for keys of the format /x/*
func (c *Client) buildNestedSecretsSlice() ([]string, error) {
	secrets := make([]string, 0)
	param := &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int64(100),
	}
	resp, err := c.client.ListSecrets(param)
	if err != nil {
		return secrets, err
	}

	for _, element := range resp.SecretList {
		secrets = append(secrets, *element.Name)
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
