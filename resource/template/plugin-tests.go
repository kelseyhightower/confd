package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/memkv"
)

const (
	tomlPath = "test/confd/config.toml"
)

// PluginTest test of some plugin for template processing
type PluginTest struct {
	Desc        string // description of the test (for helpful errors)
	Toml        string // toml file contents
	Tmpl        string // template file contents
	Expected    string // expected generated file contents
	UpdateStore func(tr memkv.Store)
}

// TestPlugin run the suite test provided by
// the plugin using the PluginTest struct
func TestPlugin(tests []PluginTest, t *testing.T) {
	for _, tt := range tests {
		executeTestPlugin(tt, t)
	}
}

// executeTestPlugin builds a TemplateResource based on the toml and tmpl files described
// in the PluginTest, writes a config file, and compares the result against the expectation
// in the PluginTest.
func executeTestPlugin(tt PluginTest, t *testing.T) {
	setupDirectoriesAndFilesForPlugins(tt, t)
	defer os.RemoveAll("test")

	tr, err := newTemplateResource()
	if err != nil {
		t.Errorf(tt.Desc + ": failed to create TemplateResource: " + err.Error())
	}

	tt.UpdateStore(tr.store)

	if err := tr.createStageFile(); err != nil {
		t.Errorf(tt.Desc + ": failed createStageFile: " + err.Error())
	}

	actual, err := ioutil.ReadFile(tr.StageFile.Name())
	if err != nil {
		t.Errorf(tt.Desc + ": failed to read StageFile: " + err.Error())
	}
	if string(actual) != tt.Expected {
		t.Errorf(fmt.Sprintf("%v: invalid StageFile. Expected %v, actual %v", tt.Desc, tt.Expected, string(actual)))
	}
}

// setupDirectoriesAndFilesForPlugins creates folders for the toml, tmpl, and output files and
// creates the toml and tmpl files as specified in the PluginTest struct.
func setupDirectoriesAndFilesForPlugins(tt PluginTest, t *testing.T) {
	// create confd directory and toml file
	if err := os.MkdirAll("./test/confd", os.ModePerm); err != nil {
		t.Errorf(tt.Desc + ": failed to created confd directory: " + err.Error())
	}
	if err := ioutil.WriteFile(tomlPath, []byte(tt.Toml), os.ModePerm); err != nil {
		t.Errorf(tt.Desc + ": failed to write toml file: " + err.Error())
	}
	// create templates directory and tmpl file
	if err := os.MkdirAll("./test/templates", os.ModePerm); err != nil {
		t.Errorf(tt.Desc + ": failed to create template directory: " + err.Error())
	}
	if err := ioutil.WriteFile("test/templates/test.conf.tmpl", []byte(tt.Tmpl), os.ModePerm); err != nil {
		t.Errorf(tt.Desc + ": failed to write toml file: " + err.Error())
	}
	// create tmp directory for output
	if err := os.MkdirAll("./test/tmp", os.ModePerm); err != nil {
		t.Errorf(tt.Desc + ": failed to create tmp directory: " + err.Error())
	}
}

// newTemplateResource creates a TemplateResource for creating a config file
func newTemplateResource() (*TemplateResource, error) {
	backendConf := backends.Config{
		Backend: "env"}
	client, err := backends.New(backendConf)
	if err != nil {
		return nil, err
	}

	config := Config{
		StoreClient: client, // not used but must be set
		TemplateDir: "./test/templates",
	}

	tr, err := NewTemplateResource(tomlPath, config)
	if err != nil {
		return nil, err
	}
	tr.Dest = "./test/tmp/test.conf"
	tr.FileMode = 0666
	return tr, nil
}
