package template

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"text/template"
	"time"

	"github.com/kelseyhightower/confd/backends/env"
	"github.com/kelseyhightower/confd/log"
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
src = "{{.src}}"
dest = "{{.dest}}"
keys = [
  "foo",
]
`

// getRandomGroup randomly selects a gid from the Groups the current user is in
func getRandomGroup() (gid int, err error) {
	ogid := os.Getgid()
	gid = ogid

	// pick a random group and set the file to that group
	groups, err := os.Getgroups()
	if err != nil {
		return
	}

	// It is possible this never terminates assert the length is greater than 1
	if len(groups) > 1 {
		rand.Seed(time.Now().UnixNano())
		for gid == ogid {
			gid = groups[rand.Intn(len(groups))]
		}
	}
	return
}

func TestProcessTemplateResources(t *testing.T) {
	log.SetLevel("warn")
	// Setup temporary conf, config, and template directories.
	tempConfDir, err := createTempDirs()
	if err != nil {
		t.Errorf("Failed to create temp dirs: %s", err.Error())
	}
	defer os.RemoveAll(tempConfDir)

	// Create the src template.
	srcTemplateFile := filepath.Join(tempConfDir, "templates", "foo.tmpl")
	err = ioutil.WriteFile(srcTemplateFile, []byte(`foo = {{getv "/foo"}}`), 0644)
	if err != nil {
		t.Error(err.Error())
	}

	// Create the dest.
	destFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Errorf("Failed to create destFile: %s", err.Error())
	}
	defer os.Remove(destFile.Name())

	gid, err := getRandomGroup()
	if err != nil {
		t.Error(err.Error())
	}
	destFile.Chown(os.Getuid(), gid)

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

	os.Setenv("FOO", "bar")
	storeClient, err := env.NewEnvClient()
	if err != nil {
		t.Errorf(err.Error())
	}
	c := Config{
		ConfDir:     tempConfDir,
		ConfigDir:   filepath.Join(tempConfDir, "conf.d"),
		StoreClient: storeClient,
		TemplateDir: filepath.Join(tempConfDir, "templates"),
	}
	// Process the test template resource.
	err = Process(c)
	if err != nil {
		t.Error(err.Error())
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

	// test the gid/uid
	destFi, err := destFile.Stat()
	if err != nil {
		t.Errorf("Issue running stat on file after create: %s", err.Error())
	} else {
		// this set of tests only works on unix like systems atm
		unixStat, ok := destFi.Sys().(*syscall.Stat_t)
		if ok {
			if int(unixStat.Uid) != os.Getuid() {
				t.Errorf("Expected template uid to be: 0 Got: %d", unixStat.Uid)
			}
			if int(unixStat.Gid) != gid {
				t.Errorf("Expected template gid to be: 100 Got: %d", unixStat.Gid)
			}
		}
	}
}

func TestSameConfigTrue(t *testing.T) {
	log.SetLevel("warn")
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
	log.SetLevel("warn")
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
