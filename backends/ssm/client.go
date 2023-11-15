package ssm

import (
	"os"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/kelseyhightower/confd/log"
)

type Client struct {
	client *ssm.Client
}

func New() (*Client, error) {
	var region string

	// get region from metadata service unless provided by env
	if os.Getenv("AWS_REGION") != "" {
		region = os.Getenv("AWS_REGION")
	} else {
		cfg, err := config.LoadDefaultConfig(context.TODO())

		if err != nil {
			return nil, err
		}

		imds_client := imds.NewFromConfig(cfg)
		response, err := imds_client.GetRegion(context.TODO(), &imds.GetRegionInput{})
		if err != nil {
			return nil, err
		}
		region = response.Region
	}

	// Create the service's client with the config.
	ssm_cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	// if SSM_LOCAL is set override the endpoint configuration
	ssm_client := ssm.NewFromConfig(ssm_cfg, func (o *ssm.Options) {
		if os.Getenv("SSM_LOCAL") != "" {
			log.Debug("SSM_LOCAL is set")
			o.BaseEndpoint = aws.String("http://localhost:8001/")
		}
	})
	return &Client{ssm_client}, nil
}

// GetValues retrieves the values for the given keys from AWS SSM Parameter Store
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	var err error
	for _, key := range keys {
		log.Debug("Processing key=%s", key)
		var resp map[string]string
		resp, err = c.getParametersWithPrefix(key)
		if err != nil {
			return vars, err
		}
		if len(resp) == 0 {
			resp, err = c.getParameter(key)
			if err != nil {
				return vars, err
			}
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
	params := &ssm.GetParametersByPathInput{
		Path:           aws.String(prefix),
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(true),
	}
	paginator := ssm.NewGetParametersByPathPaginator(c.client, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return parameters, err
		}
		for _, p := range page.Parameters {
			parameters[*p.Name] = *p.Value
		}
	}
	return parameters, err
}

func (c *Client) getParameter(name string) (map[string]string, error) {
	parameters := make(map[string]string)
	params := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	}
	resp, err := c.client.GetParameter(context.TODO(), params)
	if err != nil {
		return parameters, err
	}
	parameters[*resp.Parameter.Name] = *resp.Parameter.Value
	return parameters, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
