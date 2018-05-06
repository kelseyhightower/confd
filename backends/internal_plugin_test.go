package backends

import (
	"testing"

	"github.com/kelseyhightower/confd/plugin"
)

func TestInternalPlugin_InternalDatabases(t *testing.T) {
	for _, name := range []string{
		"env",
		"consul",
		"dynamodb",
		"etcd",
		"etcdv3",
		"rancher",
		"redis",
		"zookeeper",
		"vault",
		"file",
		"ssm",
	} {
		if _, ok := InternalDatabases[name]; !ok {
			t.Errorf("Expected to find %s in InternalDatabases", name)
		}
	}
}

func TestInternalPlugin_BuildPluginCommandString(t *testing.T) {
	actual, err := BuildPluginCommandString(plugin.DatabasePluginName, "env")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "-CONFDSPACE-internal-plugin-CONFDSPACE-database-CONFDSPACE-env"
	if actual[len(actual)-len(expected):] != expected {
		t.Errorf("Expected command to end with %s; got:\n%s\n", expected, actual)
	}
}
