package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kelseyhightower/confd/backends"
)

const (
	tomlFilePath = "test/confd/config.toml"
	tmplFilePath = "test/templates/test.conf.tmpl"
	keyFilePath  = "test/secret.gpg"
	keyContent   = `
    -----BEGIN PGP PRIVATE KEY BLOCK-----
    Version: GnuPG v2

    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
    qZMdm1eQmKhN
    =eJtp
    -----END PGP PRIVATE KEY BLOCK-----
    `
)

type templateTest struct {
	desc        string                  // description of the test (for helpful errors)
	toml        string                  // toml file contents
	tmpl        string                  // template file contents
	expected    string                  // expected generated file contents
	updateStore func(*TemplateResource) // function for setting values in store
}

// templateTests is an array of templateTest structs, each representing a test of
// some aspect of template processing. When the input tmpl and toml files are
// processed, they should produce a config file matching expected.
var templateTests = []templateTest{

	templateTest{
		desc: "base, get test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{.Value}}
{{end}}
`,
		expected: `

key: key
val: abc

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "abc")
		},
	},

	templateTest{
		desc: "gets test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/user",
    "/test/pass",
    "/nada/url",
]
`,
		tmpl: `
{{range gets "/test/*"}}
key: {{.Key}}
val: {{.Value}}
{{end}}
`,
		expected: `

key: /test/pass
val: abc

key: /test/user
val: mary

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/user", "mary")
			tr.store.Set("/test/pass", "abc")
			tr.store.Set("/nada/url", "url")
		},
	},

	templateTest{
		desc: "getv test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/url",
    "/test/user",
]
`,
		tmpl: `
url = {{getv "/test/url"}}
user = {{getv "/test/user"}}
`,
		expected: `
url = http://www.abc.com
user = bob
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/url", "http://www.abc.com")
			tr.store.Set("/test/user", "bob")
		},
	},

	templateTest{
		desc: "getvs test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/user",
    "/test/pass",
    "/nada/url",
]
`,
		tmpl: `
{{range getvs "/test/*"}}
val: {{.}}
{{end}}
`,
		expected: `

val: abc

val: mary

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/user", "mary")
			tr.store.Set("/test/pass", "abc")
			tr.store.Set("/nada/url", "url")
		},
	},

	templateTest{
		desc: "split test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data",
]
`,
		tmpl: `
{{$data := split (getv "/test/data") ":"}}
f: {{index $data 0}}
br: {{index $data 1}}
bz: {{index $data 2}}
`,
		expected: `

f: foo
br: bar
bz: baz
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", "foo:bar:baz")
		},
	},

	templateTest{
		desc: "toUpper test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/",
]
`,
		tmpl: `
{{$data := toUpper (getv "/test/data") }}
key: {{$data}}
`,
		expected: `

key: VALUE
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", `Value`)
		},
	},

	templateTest{
		desc: "toLower test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/",
]
`,
		tmpl: `
{{$data := toLower (getv "/test/data") }}
key: {{$data}}
`,
		expected: `

key: value
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", `Value`)
		},
	},

	templateTest{
		desc: "json test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/",
]
`,
		tmpl: `
{{range getvs "/test/data/*"}}
{{$data := json .}}
id: {{$data.Id}}
ip: {{$data.IP}}
{{end}}
`,
		expected: `


id: host1
ip: 192.168.10.11


id: host2
ip: 192.168.10.12

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data/1", `{"Id":"host1", "IP":"192.168.10.11"}`)
			tr.store.Set("/test/data/2", `{"Id":"host2", "IP":"192.168.10.12"}`)
		},
	},

	templateTest{
		desc: "jsonArray test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/",
]
`,
		tmpl: `
{{range jsonArray (getv "/test/data/")}}
num: {{.}}
{{end}}
`,
		expected: `

num: 1

num: 2

num: 3

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data/", `["1", "2", "3"]`)
		},
	},

	templateTest{
		desc: "ls test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/abc",
    "/test/data/def",
    "/test/data/ghi",
]
`,
		tmpl: `
{{range ls "/test/data"}}
value: {{.}}
{{end}}
`,
		expected: `

value: abc

value: def

value: ghi

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data/abc", "123")
			tr.store.Set("/test/data/def", "456")
			tr.store.Set("/test/data/ghi", "789")
		},
	},

	templateTest{
		desc: "lsdir test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/abc",
    "/test/data/def/ghi",
    "/test/data/jkl/mno",
]
`,
		tmpl: `
{{range lsdir "/test/data"}}
value: {{.}}
{{end}}
`,
		expected: `

value: def

value: jkl

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data/abc", "123")
			tr.store.Set("/test/data/def/ghi", "456")
			tr.store.Set("/test/data/jkl/mno", "789")
		},
	},
	templateTest{
		desc: "dir test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data",
    "/test/data/abc",
]
`,
		tmpl: `
{{with dir "/test/data/abc"}}
dir: {{.}}
{{end}}
`,
		expected: `

dir: /test/data

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", "parent")
			tr.store.Set("/test/data/def", "child")
		},
	},
	templateTest{
		desc: "ip lookup test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data",
    "/test/data/abc",
]
`,
		tmpl: `
{{range lookupIP "localhost"}}
ip: {{.}}
{{end}}
`,
		expected: `

ip: 127.0.0.1

ip: ::1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", "parent")
			tr.store.Set("/test/data/def", "child")
		},
	},
	templateTest{
		desc: "Decrypt test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/",
]
`,
		tmpl: `
{{$data := decrypt (getv "/test/data") "./test/secret.gpg" }}
key: {{$data}}
`,
		expected: `

key: secret
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", `hQEMA18t6iQACIUUAQf/RecRPDmzp5pIY0N7j6lXxtw+8mubIsZBo1efa1REF6fhExRg7VzGPH7qLakJFv9IHD3Sh7R+6rGRzVBHsvGAHWaxSc3fp9ftZslv7CrO4llxzd65KX9Yj842eCtG+PvgggpLvkl9p0hES1HloP/yIj+f7QfhPBbEFx40kzMCiGIsSZ3itZmFzePfbGugqPeVIPSqtaynXvyuOPhFcaguDzJCPrk2QSOPpy7gBcnF+mZ7A7vlSlsFT4nUSj0jJHzZLBfJyNbAm9tPzS8XyGQZEcMMX8W+FnQaIj0m/uWRCXQW2Z/0yJdaUbPwkzjvBsc/hyd9VjS59sy3RxGmfxpTedJMAZrsVYGaInjCbjoqy/INtEKKxXEaJcbQGGMyUigxLL0tD2IwgNO2oScZ97dRji/RJxzAmQBHENtk7W25GBSQh7iysZDi29EK8YJiOA==`)
			ioutil.WriteFile(keyFilePath, []byte(keyContent), os.ModePerm)
		},
	},
}

// TestTemplates runs all tests in templateTests
func TestTemplates(t *testing.T) {
	for _, tt := range templateTests {
		ExecuteTestTemplate(tt, t)
	}
}

// ExectureTestTemplate builds a TemplateResource based on the toml and tmpl files described
// in the templateTest, writes a config file, and compares the result against the expectation
// in the templateTest.
func ExecuteTestTemplate(tt templateTest, t *testing.T) {
	setupDirectoriesAndFiles(tt, t)
	defer os.RemoveAll("test")

	tr, err := templateResource()
	if err != nil {
		t.Errorf(tt.desc + ": failed to create TemplateResource: " + err.Error())
	}

	tt.updateStore(tr)

	if err := tr.createStageFile(); err != nil {
		t.Errorf(tt.desc + ": failed createStageFile: " + err.Error())
	}

	actual, err := ioutil.ReadFile(tr.StageFile.Name())
	if err != nil {
		t.Errorf(tt.desc + ": failed to read StageFile: " + err.Error())
	}
	if string(actual) != tt.expected {
		t.Errorf(fmt.Sprintf("%v: invalid StageFile. Expected %v, actual %v", tt.desc, tt.expected, string(actual)))
	}
}

// setUpDirectoriesAndFiles creates folders for the toml, tmpl, and output files and
// creates the toml and tmpl files as specified in the templateTest struct.
func setupDirectoriesAndFiles(tt templateTest, t *testing.T) {
	// create confd directory and toml file
	if err := os.MkdirAll("./test/confd", os.ModePerm); err != nil {
		t.Errorf(tt.desc + ": failed to created confd directory: " + err.Error())
	}
	if err := ioutil.WriteFile(tomlFilePath, []byte(tt.toml), os.ModePerm); err != nil {
		t.Errorf(tt.desc + ": failed to write toml file: " + err.Error())
	}
	// create templates directory and tmpl file
	if err := os.MkdirAll("./test/templates", os.ModePerm); err != nil {
		t.Errorf(tt.desc + ": failed to create template directory: " + err.Error())
	}
	if err := ioutil.WriteFile(tmplFilePath, []byte(tt.tmpl), os.ModePerm); err != nil {
		t.Errorf(tt.desc + ": failed to write toml file: " + err.Error())
	}
	// create tmp directory for output
	if err := os.MkdirAll("./test/tmp", os.ModePerm); err != nil {
		t.Errorf(tt.desc + ": failed to create tmp directory: " + err.Error())
	}
}

// templateResource creates a templateResource for creating a config file
func templateResource() (*TemplateResource, error) {
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

	tr, err := NewTemplateResource(tomlFilePath, config)
	if err != nil {
		return nil, err
	}
	tr.Dest = "./test/tmp/test.conf"
	tr.FileMode = 0666
	return tr, nil
}
