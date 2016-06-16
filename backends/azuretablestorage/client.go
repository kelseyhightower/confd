package azuretablestorage

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/kelseyhightower/confd/log"
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
	log.Debug(fmt.Sprintf("Query: %#v", query))
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
	// bin/confd -onetime -backend azuretablestorage -noop -confdir ~/etc/confd/ -log-level debug -table confdtest -storage-account devbnntemp -client-key xxxxx==
	log.Debug(fmt.Sprintf("table: %#v", tableName))
	log.Debug(fmt.Sprintf("account: %#v", account))
	log.Debug(fmt.Sprintf("key: %#v", key))

	client, err := storage.NewBasicClient(account, key)

	if err != nil {
		return nil, err
	}

	table := storage.AzureTable(tableName)

	return &Client{&client, table}, nil
}

// GetValues queries the environment for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {

	var te = reflect.TypeOf((*tableEntry)(nil))
	log.Debug(fmt.Sprintf("Type: %#v", te))

	vars := make(map[string]string)
	t := c.client.GetTableService()

	for _, key := range keys {
		startsWithPattern := pipeReplacer.Replace(key)
		var entities, cToken, err = queryTableRowKeyStartsWith(&t, c.table, nil, te, 1000, startsWithPattern)
		log.Debug(fmt.Sprintf("(Key: %v)Entities: %#v", startsWithPattern, entities))
		log.Debug(fmt.Sprintf("CToken: %#v", cToken))
		if err != nil {
			return vars, err
		}
		for _, e := range entities {
			en := e.(*tableEntry)
			ek := slashReplacer.Replace(en.RowKey())
			log.Debug(fmt.Sprintf("Entity: %#v", en))
			vars[ek] = en.Value
		}
	}

	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
