// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package template

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
)

type MockKey struct {
	Key   string
	Value string
}

type MockStore struct {
	Keys []*MockKey
}

func (m *MockStore) GetValues(keys []string) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	for _, key := range keys {
		for _, set := range m.Keys {
			if strings.HasPrefix(set.Key, key) {
				vals[set.Key] = set.Value
			}
		}
	}
	return vals, nil
}

func (m *MockStore) AddKey(key string, value string) {
	m.Keys = append(m.Keys, &MockKey{key, value})
}

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

var fakeFile = "/this/shoud/not/exist"

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

var templateResourceConfigWithPrefixTmpl = `
[template]
prefix = "/template_prefix"
src = "{{ .src }}"
dest = "{{ .dest }}"
keys = [
  "/foo",
]
`

func TestProcessTemplateResources(t *testing.T) {
	log.SetQuiet(true)
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
	if err := config.LoadConfig(""); err != nil {
		t.Errorf(err.Error())
	}
	config.SetPrefix("")
	// Use the temporary tempConfDir from above.
	config.SetConfDir(tempConfDir)

	// Create the stub etcd client.
	c := &MockStore{}
	c.AddKey("/foo", "bar")

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

func TestProcessTemplateResourcesWithPrefix(t *testing.T) {
	log.SetQuiet(true)
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
	tmpl, err := template.New("templateResourceConfigWithPrefixTmpl").Parse(templateResourceConfigWithPrefixTmpl)
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
	if err := config.LoadConfig(""); err != nil {
		t.Errorf(err.Error())
	}
	config.SetPrefix("")
	// Use the temporary tempConfDir from above.
	config.SetConfDir(tempConfDir)

	// Create the stub etcd client.
	c := &MockStore{}
	c.AddKey("/template_prefix/foo", "bar")

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

func TestProcessTemplateResourcesNoop(t *testing.T) {
	log.SetQuiet(true)
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
	if err := config.LoadConfig(""); err != nil {
		t.Errorf(err.Error())
	}
	config.SetPrefix("")
	// Use the temporary tempConfDir from above.
	config.SetConfDir(tempConfDir)
	// Enable noop mode.
	config.SetNoop(true)

	// Create the stub etcd client.
	c := &MockStore{}
	c.AddKey("/foo", "bar")

	// Process the test template resource.
	runErrors := ProcessTemplateResources(c)
	if len(runErrors) > 0 {
		for _, e := range runErrors {
			t.Errorf(e.Error())
		}
	}
	// Verify the results.
	expected := ""
	results, err := ioutil.ReadFile(destFile.Name())
	if err != nil {
		t.Error(err.Error())
	}
	if string(results) != expected {
		t.Errorf("Expected contents of dest == '%s', got %s", expected, string(results))
	}
}

func TestBrokenTemplateResourceFile(t *testing.T) {
	log.SetQuiet(true)
	tempFile, err := ioutil.TempFile("", "")
	defer os.Remove(tempFile.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = tempFile.WriteString(brokenTemplateResourceConfig)
	if err != nil {
		t.Errorf(err.Error())
	}
	// Create the stub client.
	c := &MockStore{}

	// Process broken template resource config file.
	_, err = NewTemplateResourceFromPath(tempFile.Name(), c)
	if err == nil {
		t.Errorf("Expected err not to be nil")
	}
}

func TestSameConfigTrue(t *testing.T) {
	log.SetQuiet(true)
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
	log.SetQuiet(true)
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

func TestIsFileExist(t *testing.T) {
	log.SetQuiet(true)
	result := isFileExist(fakeFile)
	if result != false {
		t.Errorf("Expected IsFileExist(%s) to be false, got %v", fakeFile, result)
	}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(f.Name())
	result = isFileExist(f.Name())
	if result != true {
		t.Errorf("Expected IsFileExist(%s) to be true, got %v", f.Name(), result)
	}
}

type PathToKeyTest struct {
	key, prefix, expected string
}

var pathToKeyTests = []PathToKeyTest{
	// Without prefix
	{"/nginx/port", "", "nginx_port"},
	{"/nginx/worker_processes", "", "nginx_worker_processes"},
	{"/foo/bar/mat/zoo", "", "foo_bar_mat_zoo"},
	// With prefix
	{"/prefix/nginx/port", "/prefix", "nginx_port"},
	// With prefix and trailing slash
	{"/prefix/nginx/port", "/prefix/", "nginx_port"},
}

func TestPathToKey(t *testing.T) {
	for _, pt := range pathToKeyTests {
		result := pathToKey(pt.key, pt.prefix)
		if result != pt.expected {
			t.Errorf("Expected pathToKey(%s, %s) to == %s, got %s",
				pt.key, pt.prefix, pt.expected, result)
		}
	}
}

func TestCleanKeys(t *testing.T) {
	pre := map[string]interface{}{
		"/my/cool_val/here":     "test",
		"/this_key":             "foo",
		"/prefix/key":           "test",
		"/my/new-cool-val/here": "bar",
		"prefix/noslash":        "test",
	}
	config.SetPrefix("/prefix")
	clean := cleanKeys(pre, "/prefix")
	if len(clean) != len(pre) {
		t.Fatalf("bad length")
	}
	if _, ok := clean["my_cool_val_here"]; !ok {
		t.Fatalf("bad: %v", clean)
	}
	if _, ok := clean["this_key"]; !ok {
		t.Fatalf("bad: %v", clean)
	}
	if _, ok := clean["key"]; !ok {
		t.Fatalf("bad: %v", clean)
	}
	if _, ok := clean["my_new_cool_val_here"]; !ok {
		t.Fatalf("bad: %v", clean)
	}
	if _, ok := clean["noslash"]; !ok {
		t.Fatalf("bad: %v", clean)
	}
}
