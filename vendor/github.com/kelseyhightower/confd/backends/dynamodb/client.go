package dynamodb

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/kelseyhightower/confd/log"
)

// Client is a wrapper around the DynamoDB client
// and also holds the table to lookup key value pairs from
type Client struct {
	client *dynamodb.DynamoDB
	table  string
}

// NewDynamoDBClient returns an *dynamodb.Client with a connection to the region
// configured via the AWS_REGION environment variable.
// It returns an error if the connection cannot be made or the table does not exist.
func NewDynamoDBClient(table string) (*Client, error) {
	var c *aws.Config
	if os.Getenv("DYNAMODB_LOCAL") != "" {
		log.Debug("DYNAMODB_LOCAL is set")
		endpoint := "http://localhost:8000"
		c = &aws.Config{
			Endpoint: &endpoint,
		}
	} else {
		c = nil
	}

	session := session.New(c)

	// Fail early, if no credentials can be found
	_, err := session.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	d := dynamodb.New(session)

	// Check if the table exists
	_, err = d.DescribeTable(&dynamodb.DescribeTableInput{TableName: &table})
	if err != nil {
		return nil, err
	}
	return &Client{d, table}, nil
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
					log.Warning("Skipping key '%s'. 'value' is not of type 'string'.", key)
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
					log.Warning("Skipping key '%s'. 'value' is not of type 'string'.", *item["key"].S)
				}
				continue
			}
		}
	}
	return vars, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
