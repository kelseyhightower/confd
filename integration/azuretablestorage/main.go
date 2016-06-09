package main

import (
	"fmt"
	"os"
	"reflect"

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

func containsTable(list []storage.AzureTable, elem storage.AzureTable) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

func main() {
	account := os.Args[1]
	key := os.Args[2]
	tableName := os.Args[3]

	client, err := storage.NewBasicClient(account, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	table := storage.AzureTable(tableName)

	ts := client.GetTableService()

	//err = ts.DeleteTable(table)

	s, err := ts.QueryTables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if containsTable(s, table) {
		fmt.Fprintf(os.Stdout, "Table %v already exists, using existing table\n", tableName)
	} else {
		fmt.Fprintf(os.Stdout, "Creating test table: %v\n", tableName)
		err = ts.CreateTable(table)
	}

	var te = reflect.TypeOf((*tableEntry)(nil))
	existingEntities, _, err := ts.QueryTableEntities(table, nil, te, 1000, "")

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, e := range existingEntities {
		fmt.Fprintf(os.Stdout, "Deleting old key: %v\n", e.RowKey())
		ts.DeleteEntity(table, e, "")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Inserting test data into table: %v\n", tableName)

	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|key", Value: "foobar"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|database|host", Value: "127.0.0.1"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|database|password", Value: "p@sSw0rd"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|database|port", Value: "3306"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|database|username", Value: "confd"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|upstream|app1", Value: "10.0.1.10:8080"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|upstream|app2", Value: "10.0.1.11:8080"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|upstream|broken", Value: "4711"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|prefix|database|host", Value: "127.0.0.1"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|prefix|database|password", Value: "p@sSw0rd"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|prefix|database|port", Value: "3306"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|prefix|database|username", Value: "confd"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|prefix|upstream|app1", Value: "0.0.1.10:8080"})
	ts.InsertEntity(table, &tableEntry{partitionKey: "1", rowKey: "|prefix|upstream|app2", Value: "10.0.1.11:8080"})
}
