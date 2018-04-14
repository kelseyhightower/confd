package secretsmanager

//see https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/kelseyhightower/confd/log"
)

type Client struct {
	client *secretsmanager.SecretsManager
}

func New() (*Client, error) {
	log.Debug("creating secretsmanager client")
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
		endpoint := "http://localhost:8001"
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
	var err error
	for _, key := range keys {
		log.Debug("Processing key=%s", key)
		var resp map[string]string
		resp, err = c.getSecretValue(key)
		if err != nil {
			return vars, err
		}
		for k, v := range resp {
			vars[k] = v
		}
	}
	return vars, nil
}

func (c *Client) getSecretValue(name string) (map[string]string, error) {
	secret := make(map[string]string)
	params := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	}
	resp, err := c.client.GetSecretValue(params)
	if err != nil {
		return secret, err
	}
	secret[*resp.Name] = *resp.SecretString
	return secret, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
