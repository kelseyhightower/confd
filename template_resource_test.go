package main

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/confd/etcdtest"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"text/template"
)

// createTempDirs is a helper function which creates temporary directories
// required by confd. createTempDirs returns the path name representing the
// confd confDir.
// It returns an error if any.
func createTempDirs() (string, error) {
	confDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	err = os.Mkdir(filepath.Join(confDir, "templates"), 0755)
	if err != nil {
		return "", err
	}
	err = os.Mkdir(filepath.Join(confDir, "conf.d"), 0755)
	if err != nil {
		return "", err
	}
	return confDir, nil
}

var templateResourceConfigTmpl = `
[template]
src = "{{ .src }}"
dest = "{{ .dest }}"
keys = [
  "/foo",
]
`

var brokenTemplateResourceConfig = `
[template]
src = "/does/not/exist"
dest = "/does/not/exist"
keys = [
  "/foo"
  "/bar"
]
`

func TestProcessTemplateResources(t *testing.T) {
	// Setup temporary conf, config, and template directories.
	tempConfDir, err := createTempDirs()
	if err != nil {
		t.Errorf("Failed to create temp dirs: %s", err.Error())
	}
	defer os.RemoveAll(tempConfDir)

	// Create the src template.
	srcTemplateFile := filepath.Join(tempConfDir, "templates", "foo.tmpl")
	err = ioutil.WriteFile(srcTemplateFile, []byte("foo = {{ .foo }}"), 0644)
	if err != nil {
		t.Error(err.Error())
	}

	// Create the dest.
	destFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Errorf("Failed to create destFile: %s", err.Error())
	}
	defer os.Remove(destFile.Name())

	// Create the template resource configuration file.
	templateResourcePath := filepath.Join(tempConfDir, "conf.d", "foo.toml")
	templateResourceFile, err := os.Create(templateResourcePath)
	if err != nil {
		t.Errorf(err.Error())
	}
	tmpl, err := template.New("templateResourceConfig").Parse(templateResourceConfigTmpl)
	if err != nil {
		t.Errorf("Unable to parse template resource template: %s", err.Error())
	}
	data := make(map[string]string)
	data["src"] = "foo.tmpl"
	data["dest"] = destFile.Name()
	err = tmpl.Execute(templateResourceFile, data)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Load the confd configuration settings.
	if err := loadConfig(""); err != nil {
		t.Errorf(err.Error())
	}
	config.Confd.Prefix = ""
	// Use the temporary tempConfDir from above.
	config.Confd.ConfDir = tempConfDir

	// Create the stub etcd client.
	c := etcdtest.NewClient()
	fooResp := []*etcd.Response{
		{Key: "/foo", Dir: false, Value: "bar"},
	}
	c.AddResponse("/foo", fooResp)

	// Process the test template resource.
	runErrors := ProcessTemplateResources(c)
	if len(runErrors) > 0 {
		for _, e := range runErrors {
			t.Errorf(e.Error())
		}
	}
	// Verify the results.
	expected := "foo = bar"
	results, err := ioutil.ReadFile(destFile.Name())
	if err != nil {
		t.Error(err.Error())
	}
	if string(results) != expected {
		t.Errorf("Expected contents of dest == '%s', got %s", expected, string(results))
	}
}

func TestBrokenTemplateResourceFile(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "")
	defer os.Remove(tempFile.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = tempFile.WriteString(brokenTemplateResourceConfig)
	if err != nil {
		t.Errorf(err.Error())
	}
	// Create the stub etcd client.
	c := etcdtest.NewClient()
	// Process broken template resource config file.
	_, err = NewTemplateResourceFromPath(tempFile.Name(), c)
	if err == nil {
		t.Errorf("Expected err not to be nil")
	}
}

func TestSameConfigTrue(t *testing.T) {
	src, err := ioutil.TempFile("", "src")
	defer os.Remove(src.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = src.WriteString("foo")
	if err != nil {
		t.Errorf(err.Error())
	}
	dest, err := ioutil.TempFile("", "dest")
	defer os.Remove(dest.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = dest.WriteString("foo")
	if err != nil {
		t.Errorf(err.Error())
	}
	status, err := sameConfig(src.Name(), dest.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	if status != true {
		t.Errorf("Expected sameConfig(src, dest) to be %v, got %v", true, status)
	}
}

func TestSameConfigFalse(t *testing.T) {
	src, err := ioutil.TempFile("", "src")
	defer os.Remove(src.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = src.WriteString("src")
	if err != nil {
		t.Errorf(err.Error())
	}
	dest, err := ioutil.TempFile("", "dest")
	defer os.Remove(dest.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = dest.WriteString("dest")
	if err != nil {
		t.Errorf(err.Error())
	}
	status, err := sameConfig(src.Name(), dest.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	if status != false {
		t.Errorf("Expected sameConfig(src, dest) to be %v, got %v", false, status)
	}
}
