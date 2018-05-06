package secretsmanager

//see https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/

// Secrets Manager does not have the equivelent of variables defined by paths eg /varroot/var1, /varoot/var2.
// Consequently we have to parse the secrets and look for '/' to emulate that functionality
// Only supports SecretString, and retreives the most current secret

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
	keysToRetrieve := make([]string, 0)
	knownkeys, err := c.buildNestedSecretsSlice()
	if err != nil {
		return vars, err
	}
	sort.Strings(knownkeys)
	sort.Strings(keys)
	log.Debug("known keys =%v \n keys =%v", knownkeys, keys)

	klen := len(keys)
	kklen := len(knownkeys)
	k := 0
	// This works because both sets of keys are sorted and we keep a reference of where we are up to so as not to access any value more than once
	for kk := 0; kk < kklen && k < klen; {
		found := false
		for kk < kklen && keyMatches(knownkeys[kk], keys[k]) {
			found = true // knownkeys entry is found with the prefix we are looking for
			keysToRetrieve = append(keysToRetrieve, knownkeys[kk])
			kk++
		}
		if kk >= kklen {
			break // there are no more knownkeys
		} else if found || keys[k] < knownkeys[kk] {
			// increment the key as there are no more known keys thats start with keys[i]
			k++
		} else {
			// go to the next known key
			kk++
		}
	}

	for _, element := range keysToRetrieve {
		resp, err := c.getSecretValue(element)
		if err != nil {
			return vars, err
		}
		vars[resp.Name] = resp.Secret
	}
	return vars, nil
}

// build a slice of keys by calling ListSecrets keys
func (c *Client) buildNestedSecretsSlice() ([]string, error) {
	secrets := make([]string, 0)
	param := &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int64(100),
	}

	err := c.client.ListSecretsPages(param,
		func(s *secretsmanager.ListSecretsOutput, lastPage bool) bool {
			for _, element := range s.SecretList {
				secrets = append(secrets, *element.Name)
			}
			return lastPage
		})

	if err != nil {
		return secrets, err
	}
	return secrets, err
}

// Retreive secret value from AWS
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

// logic for matching keys
func keyMatches(knownkey string, key string) bool {
	// exact match
	if knownkey == key {
		return true
	}
	// prep the key so we dont get partial matches
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}

	if strings.HasPrefix(knownkey, key) {
		return true
	} else {
		return false
	}
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
