package file

/*
import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/kelseyhightower/confd/log"
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

const overrideValue1 = "" +
	"---\n" +
	"database:\n" +
	"  password: ENVp@sSw0rd\n"

const overrideValue2 = "" +
	"---\n" +
	"prefix:\n" +
	"  database:\n" +
	"    password: ENVp@sSw0rd\n"

var overrideExpected = map[string]string{
	"/key":                      "foobar",
	"/database/host":            "127.0.0.1",
	"/database/password":        "ENVp@sSw0rd",
	"/database/port":            "3306",
	"/database/username":        "confd",
	"/upstream/app1":            "10.0.1.10:8080",
	"/upstream/app2":            "10.0.1.11:8080",
	"/prefix/database/host":     "127.0.0.1",
	"/prefix/database/password": "ENVp@sSw0rd",
	"/prefix/database/port":     "3306",
	"/prefix/database/username": "confd",
	"/prefix/upstream/app1":     "10.0.1.10:8080",
	"/prefix/upstream/app2":     "10.0.1.11:8080"}

type Yaml struct {
	env   bool
	file  bool
	value string
}

type YamlMeta struct {
	yaml Yaml
	name string
}

func getValues(yamls []Yaml, keys ...string) (map[string]string, error) {
	tempDir := os.TempDir()
	yamlMeta := make([]YamlMeta, len(yamls))

	for i, yaml := range yamls {
		var name string
		if yaml.file {
			name = path.Join(tempDir,
				fmt.Sprintf("CONFD_CLCONF_TEST_%d.yml", i))
		} else {
			name = fmt.Sprintf("CONFD_CLCONF_TEST_%d", i)
		}
		yamlMeta[i] = YamlMeta{yaml, name}
	}

	defer func() {
		for _, meta := range yamlMeta {
			if meta.yaml.env {
				os.Unsetenv(meta.name)
			}
		}
		os.Unsetenv("YAML_VARS")
		os.Unsetenv("YAML_FILES")
		os.RemoveAll(tempDir)
	}()

	var yamlFile, yamlBase64, envYamlFile, envYamlBase64 string
	for _, meta := range yamlMeta {
		if meta.yaml.env {
			if meta.yaml.file {
				if envYamlFile != "" {
					envYamlFile += ","
				}
				envYamlFile += meta.name
				ioutil.WriteFile(
					meta.name,
					[]byte(meta.yaml.value), 0700)
			} else {
				if envYamlBase64 != "" {
					envYamlBase64 += ","
				}
				os.Setenv(
					meta.name,
					base64.StdEncoding.EncodeToString(
						[]byte(meta.yaml.value)))
				envYamlBase64 += meta.name
			}
		} else {
			if meta.yaml.file {
				if yamlFile != "" {
					yamlFile += ","
				}
				yamlFile += meta.name
				ioutil.WriteFile(
					meta.name,
					[]byte(meta.yaml.value), 0700)
			} else {
				if yamlBase64 != "" {
					yamlBase64 += ","
				}
				yamlBase64 += base64.StdEncoding.EncodeToString(
					[]byte(meta.yaml.value))
			}
		}
	}
	if envYamlFile != "" {
		os.Setenv("YAML_FILES", envYamlFile)
	}
	if envYamlBase64 != "" {
		os.Setenv("YAML_VARS", envYamlBase64)
	}

	log.Info(fmt.Sprintf("yamlFile:[%s], yamlBase64:[%s], envYamlFile:[%s], envYamlBase64:[%s]",
		yamlFile, yamlBase64, envYamlFile, envYamlBase64))
	client, err := NewFileClient(yamlFile, yamlBase64)
	if err != nil {
		return nil, err
	}

	return client.GetValues(keys)
}

func TestGetValues(t *testing.T) {
	values, err := getValues(
		[]Yaml{
			Yaml{env: false, file: true, value: configMap},
			Yaml{env: false, file: true, value: secrets},
		},
		"/")
	if err != nil {
		t.Errorf("Failed to get values: %v", err)
	}
	if !reflect.DeepEqual(expected, values) {
		t.Errorf("Failed get values: [%v] != [%v]", expected, values)
	}
}

func TestGetValuesWithOverrides(t *testing.T) {
	values, err := getValues(
		[]Yaml{
			Yaml{env: false, file: true, value: configMap},
			Yaml{env: false, file: true, value: secrets},
			Yaml{env: false, file: false, value: overrideValue1},
			Yaml{env: false, file: false, value: overrideValue2},
		},
		"/")
	if err != nil {
		t.Errorf("Failed to get values: %v", err)
	}
	if !reflect.DeepEqual(overrideExpected, values) {
		t.Errorf("Failed get values: [%v] != [%v]", overrideExpected, values)
	}
}

func TestGetValuesFromEnvironment(t *testing.T) {
	values, err := getValues(
		[]Yaml{
			Yaml{env: true, file: true, value: configMap},
			Yaml{env: true, file: true, value: secrets},
		},
		"/")

	if err != nil {
		t.Errorf("Failed to get values: %v", err)
	}

	if !reflect.DeepEqual(expected, values) {
		t.Errorf("Failed get values: [%v] != [%v]", expected, values)
	}
}

func TestGetValuesFromEnvironmentWithOverrides(t *testing.T) {
	values, err := getValues(
		[]Yaml{
			Yaml{env: true, file: true, value: configMap},
			Yaml{env: true, file: true, value: secrets},
			Yaml{env: true, file: false, value: overrideValue1},
			Yaml{env: true, file: false, value: overrideValue2},
		},
		"/")

	if err != nil {
		t.Errorf("Failed to get values: %v", err)
	}

	if !reflect.DeepEqual(overrideExpected, values) {
		t.Errorf("Failed get values: [%v] != [%v]", overrideExpected, values)
	}
}

func TestGetValuesMixed(t *testing.T) {
	values, err := getValues(
		[]Yaml{
			Yaml{env: false, file: true, value: configMap},
			Yaml{env: true, file: true, value: secrets},
			Yaml{env: false, file: false, value: overrideValue1},
			Yaml{env: true, file: false, value: overrideValue2},
		},
		"/")

	if err != nil {
		t.Errorf("Failed to get values: %v", err)
	}

	if !reflect.DeepEqual(overrideExpected, values) {
		t.Errorf("Failed get values: [%v] != [%v]", overrideExpected, values)
	}
}
*/
