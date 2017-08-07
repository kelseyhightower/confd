package dynamodb

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mitchellh/mapstructure"
)

// Client is a wrapper around the DynamoDB client
// and also holds the table to lookup key value pairs from
type Client struct {
	client *dynamodb.DynamoDB
	table  string
}

// Configure configures *dynamodb.Client with a connection to the region
// configured via the AWS_REGION environment variable.
// It returns an error if the connection cannot be made or the table does not exist.
func (c *Client) Configure(configRaw map[string]string) error {
	var config Config
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return err
	}

	c.table = config.Table
	var awsConfig *aws.Config
	if os.Getenv("DYNAMODB_LOCAL") != "" {
		log.Println("DYNAMODB_LOCAL is set")
		endpoint := "http://localhost:8000"
		awsConfig = &aws.Config{
			Endpoint: &endpoint,
		}
	} else {
		awsConfig = nil
	}

	session := session.New(awsConfig)

	// Fail early, if no credentials can be found
	_, err := session.Config.Credentials.Get()
	if err != nil {
		return err
	}

	d := dynamodb.New(session)

	// Check if the table exists
	_, err = d.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(c.table),
	})
	if err != nil {
		return err
	}

	c.client = d
	return nil
}

// GetValues retrieves the values for the given keys from DynamoDB
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		// Check if we can find the single item
		m := make(map[string]*dynamodb.AttributeValue)
		m["key"] = &dynamodb.AttributeValue{S: aws.String(key)}
		g, err := c.client.GetItem(&dynamodb.GetItemInput{Key: m, TableName: &c.table})
		if err != nil {
			return vars, err
		}

		if g.Item != nil {
			if val, ok := g.Item["value"]; ok {
				if val.S != nil {
					vars[key] = *val.S
				} else {
					log.Printf("Skipping key '%s'. 'value' is not of type 'string'.", key)
				}
				continue
			}
		}

		// Check for nested keys
		q, err := c.client.Scan(
			&dynamodb.ScanInput{
				ScanFilter: map[string]*dynamodb.Condition{
					"key": &dynamodb.Condition{
						AttributeValueList: []*dynamodb.AttributeValue{
							&dynamodb.AttributeValue{S: aws.String(key)}},
						ComparisonOperator: aws.String("BEGINS_WITH")}},
				AttributesToGet: []*string{aws.String("key"), aws.String("value")},
				TableName:       aws.String(c.table),
				Select:          aws.String("SPECIFIC_ATTRIBUTES"),
			})

		if err != nil {
			return vars, err
		}

		for _, i := range q.Items {
			item := i
			if val, ok := item["value"]; ok {
				if val.S != nil {
					vars[*item["key"].S] = *val.S
				} else {
					log.Printf("Skipping key '%s'. 'value' is not of type 'string'.", *item["key"].S)
				}
				continue
			}
		}
	}
	return vars, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64) (uint64, error) {
	return 0, nil
}
