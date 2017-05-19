package ssm

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/kelseyhightower/confd/log"
)

type Client struct {
	client *ssm.SSM
}

func New() (*Client, error) {
	// Create a session to share configuration, and load external configuration.
	sess := session.Must(session.NewSession())

	// Fail early, if no credentials can be found
	_, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	var c *aws.Config
	if os.Getenv("SSM_LOCAL") != "" {
		log.Debug("SSM_LOCAL is set")
		endpoint := "http://localhost:8001"
		c = &aws.Config{
			Endpoint: &endpoint,
		}
	} else {
		c = nil
	}

	// Create the service's client with the session.
	svc := ssm.New(sess, c)
	return &Client{svc}, nil
}

func isKeyValid(key string) bool {
	return strings.Contains(key, "/")
}

// GetValues retrieves the values for the given keys from AWS SSM Parameter Store
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	var err error
	for _, key := range keys {
		if isKeyValid(key) == true {
			return vars, fmt.Errorf("Key=%s is invalid", key)
		}
		log.Debug("Processing key=%s", key)
		var resp map[string]string
		resp, err = c.getParametersWithPrefix(key)
		if err != nil {
			return vars, err
		}
		for k, v := range resp {
			vars[k] = v
		}
	}
	return vars, nil
}

func (c *Client) getParametersWithPrefix(prefix string) (map[string]string, error) {
	var err error
	parameters := make(map[string]string)
	params := &ssm.DescribeParametersInput{
		Filters: []*ssm.ParametersFilter{
			{
				Values: []*string{
					aws.String(prefix),
				},
				Key: aws.String("Name"),
			},
		},
		MaxResults: aws.Int64(50),
	}
	var names []string
	resp := &ssm.DescribeParametersOutput{
		NextToken: aws.String(""),
	}
	for resp.NextToken != nil {
		resp, err = c.client.DescribeParameters(params)
		if err != nil {
			return parameters, err
		}
		if resp.NextToken != nil {
			params.SetNextToken(*aws.String(*resp.NextToken))
		}
		for _, p := range resp.Parameters {
			names = append(names, *p.Name)
		}
	}
	parameters, err = c.getParametersValues(names)
	return parameters, err
}

func (c *Client) getParametersValues(names []string) (map[string]string, error) {
	parameters := make(map[string]string)
	for _, name := range names {
		params := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(name),
			},
			WithDecryption: aws.Bool(true),
		}
		resp, err := c.client.GetParameters(params)
		if err != nil {
			return parameters, err
		}
		for _, p := range resp.Parameters {
			parameters[*p.Name] = *p.Value
		}
	}
	return parameters, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
