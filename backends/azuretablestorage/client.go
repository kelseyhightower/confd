package azuretablestorage

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type tableEntry struct {
	partitionKey string
	rowKey       string
	Value        string
}

func (t tableEntry) PartitionKey() string {
	return t.partitionKey
}

func (t tableEntry) RowKey() string {
	return t.rowKey
}

func (t *tableEntry) SetPartitionKey(v string) error {
	t.partitionKey = v
	return nil
}

func (t *tableEntry) SetRowKey(v string) error {
	t.rowKey = v
	return nil
}

func queryTableRowKeyStartsWith(t *storage.TableServiceClient, tableName storage.AzureTable, previousContinuationToken *storage.ContinuationToken, retType reflect.Type, top int, startsWithPattern string) ([]storage.TableEntity, *storage.ContinuationToken, error) {
	length := len(startsWithPattern) - 1
	lastChar := startsWithPattern[length]
	nextLastChar := lastChar + 1
	startsWithEndPattern := string(startsWithPattern[:length]) + string(nextLastChar)
	query := fmt.Sprintf("RowKey ge '%v' and RowKey lt '%v'", startsWithPattern, startsWithEndPattern)
	return t.QueryTableEntities(tableName, previousContinuationToken, retType, top, query)
}

var slashReplacer = strings.NewReplacer("|", "/")
var pipeReplacer = strings.NewReplacer("/", "|")

// Client provides a shell for the azuretablestorage client
type Client struct {
	client *storage.Client
	table  storage.AzureTable
}

// NewAzureTableStorageClient returns a new client
func NewAzureTableStorageClient(tableName string, account string, key string) (*Client, error) {
	client, err := storage.NewBasicClient(account, key)
	if err != nil {
		return nil, err
	}

	table := storage.AzureTable(tableName)

	return &Client{&client, table}, nil
}

// GetValues queries the environment for keys
func (c *Client) GetValues(keys []string, token string) (map[string]string, error) {

	var te = reflect.TypeOf((*tableEntry)(nil))

	vars := make(map[string]string)
	t := c.client.GetTableService()
	var cToken *storage.ContinuationToken

	for _, key := range keys {
		startsWithPattern := pipeReplacer.Replace(key)
		for {
			entities, cToken, err := queryTableRowKeyStartsWith(&t, c.table, cToken, te, 100, startsWithPattern)
			if err != nil {
				return vars, err
			}

			for _, e := range entities {
				en := e.(*tableEntry)
				ek := slashReplacer.Replace(en.RowKey())
				vars[ek] = en.Value
			}

			if cToken == nil {
				break
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
