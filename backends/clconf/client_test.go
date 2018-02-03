package clconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

const configMap = "" +
	"---\n" +
	"key: foobar\n" +
	"database:\n" +
	"  host: 127.0.0.1\n" +
	"  port: \"3306\"\n" +
	"upstream:\n" +
	"  app1: 10.0.1.10:8080\n" +
	"  app2: 10.0.1.11:8080\n" +
	"prefix:\n" +
	"  database:\n" +
	"    host: 127.0.0.1\n" +
	"    port: \"3306\"\n" +
	"  upstream:\n" +
	"    app1: 10.0.1.10:8080\n" +
	"    app2: 10.0.1.11:8080\n"

const secrets = "" +
	"---\n" +
	"database:\n" +
	"  password: p@sSw0rd\n" +
	"  username: confd\n" +
	"prefix:\n" +
	"  database:\n" +
	"    password: p@sSw0rd\n" +
	"    username: confd\n"

var expected = map[string]string{
	"/key":                      "foobar",
	"/database/host":            "127.0.0.1",
	"/database/password":        "p@sSw0rd",
	"/database/port":            "3306",
	"/database/username":        "confd",
	"/upstream/app1":            "10.0.1.10:8080",
	"/upstream/app2":            "10.0.1.11:8080",
	"/prefix/database/host":     "127.0.0.1",
	"/prefix/database/password": "p@sSw0rd",
	"/prefix/database/port":     "3306",
	"/prefix/database/username": "confd",
	"/prefix/upstream/app1":     "10.0.1.10:8080",
	"/prefix/upstream/app2":     "10.0.1.11:8080"}

func getValues(keys ...string) (map[string]string, error) {
	tempDir := os.TempDir()
	defer func() {
		os.RemoveAll(tempDir)
	}()

	configMapFile := path.Join(tempDir, "configMap.yml")
	ioutil.WriteFile(configMapFile, []byte(configMap), 0700)
	secretsFile := path.Join(tempDir, "secrets.yml")
	ioutil.WriteFile(secretsFile, []byte(secrets), 0700)

	yamlFiles := fmt.Sprintf("%s,%s", configMapFile, secretsFile)
	fmt.Printf("REMOVE ME: yamlFiles [%s]\n", yamlFiles)
	client, err := NewClconfClient(yamlFiles)
	if err != nil {
		return nil, err
	}

	return client.GetValues(keys)
}

func TestGetValues(t *testing.T) {
	values, err := getValues("/")
	if err != nil {
		t.Errorf("Failed to get values: %v", err)
	}

	if !reflect.DeepEqual(expected, values) {
		t.Errorf("Failed get values: [%v] != [%v]", expected, values)
	}
}
