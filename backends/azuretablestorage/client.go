package azuretablestorage

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/kelseyhightower/confd/log"
)

// Client provides a shell for the azuretablestorage client
type Client struct {
	client *storage.Client
	table  storage.AzureTable
}

type TableEntry struct {
	partitionKey string
	rowKey       string
	Value        string
}

func (t TableEntry) PartitionKey() string {
	return t.partitionKey
}

func (t TableEntry) RowKey() string {
	return t.rowKey
}

func (t *TableEntry) SetPartitionKey(v string) error {
	t.partitionKey = v
	return nil
}

func (t *TableEntry) SetRowKey(v string) error {
	t.rowKey = v
	return nil
}

var replacer = strings.NewReplacer("-", "/")

// NewAzureTableStorageClient returns a new client
func NewAzureTableStorageClient() (*Client, error) {
	client, err := storage.NewBasicClient("devbnntemp", "YXbC2OZEJK17BTjEz6hMUpHzyegqZeHFi+DVIrP7simQURu12YYJMaj7vnQbxLGe8JxKzgShXPH9R93GrZCmFA==")

	if err != nil {
		return nil, err
	}

	table := storage.AzureTable("confdtest")

	return &Client{&client, table}, nil
}

// GetValues queries the environment for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {

	var te = reflect.TypeOf((*TableEntry)(nil))
	log.Debug(fmt.Sprintf("Type: %#v", te))

	vars := make(map[string]string)
	t := c.client.GetTableService()
	//var cToken storage.ContinuationToken

	var entities, cToken, err = t.QueryTableEntities(c.table, nil, te, 100, "")

	log.Debug(fmt.Sprintf("Entities: %#v", entities))
	log.Debug(fmt.Sprintf("CToken: %#v", cToken))

	if err != nil {
		return vars, err
	}

	for _, e := range entities {
		en := e.(*TableEntry)
		k := replacer.Replace(en.RowKey())
		log.Debug(fmt.Sprintf("Entity: %#v", en))
		vars[k] = en.Value
	}

	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
