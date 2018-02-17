package clconf

import (
	"encoding/base64"
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/urfave/cli"
)

func NewTestConfig() (interface{}, error) {
	config, err := NewTestConfigContent()
	if err != nil {
		return "", err
	}
	return unmarshalYaml(config)
}

func NewTestConfigBase64() (string, error) {
	config, err := NewTestConfigContent()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(config)), nil
}

func NewTestConfigContent() ([]byte, error) {
	return ioutil.ReadFile(NewTestConfigFile())
}

func NewTestConfigFile() string {
	return filepath.Join("testdata", "testconfig.yml")
}

func NewTestContext(name string, app *cli.App, flags []cli.Flag, parentContext *cli.Context, args []string, options []string) *cli.Context {
	set := flag.NewFlagSet(name, 0)
	for _, flag := range flags {
		flag.Apply(set)
	}
	set.SetOutput(ioutil.Discard)
	context := cli.NewContext(app, set, parentContext)
	set.Parse(append(options, args...))
	return context
}

func NewTestGetvContext(args []string, options []string) *cli.Context {
	context := NewTestContext("getv", nil, getvFlags(), NewTestGlobalContext(), args, options)
	return context
}

func NewTestGlobalContext() *cli.Context {
	return NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{
		"--secret-keyring", NewTestKeysFile(),
		"--yaml", NewTestConfigFile(),
	})
}

func NewTestKeysFile() string {
	return filepath.Join("testdata", "test.secring.gpg")
}

func NewTestSecretAgent() (*SecretAgent, error) {
	return NewSecretAgentFromFile(NewTestKeysFile())
}

func NewTestSetvContext(yamlFile string, args []string, options []string) *cli.Context {
	globalContext := NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{
		"--secret-keyring", NewTestKeysFile(),
		"--yaml", yamlFile,
	})
	return NewTestContext("setv", nil, setvFlags(), globalContext, args, options)
}

func ValuesAtPathsAreEqual(config interface{}, a, b string) bool {
	aValue, ok := GetValue(config, a)
	if !ok {
		return false
	}
	bValue, ok := GetValue(config, b)
	if !ok {
		return false
	}
	return aValue == bValue
}
