package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/abtreece/confd/pkg/backends"
)

const (
	tomlFilePath = "test/confd/config.toml"
	tmplFilePath = "test/templates/test.conf.tmpl"
)

type templateTest struct {
	desc        string                  // description of the test (for helpful errors)
	toml        string                  // toml file contents
	tmpl        string                  // template file contents
	expected    interface{}             // expected generated file contents
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
		desc: "ipv4 lookup test",
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
{{range lookupIPV4 "localhost"}}
ip: {{.}}
{{end}}
`,
		expected: `

ip: 127.0.0.1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", "parent")
			tr.store.Set("/test/data/def", "child")
		},
	},
	templateTest{
		desc: "ipv6 lookup test",
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
{{range lookupIPV6 "localhost"}}
ip: {{.}}
{{end}}
`,
		expected: [...]string{
			`
ip: ::1

`,
			`

`,
		},
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
		expected: [...]string{
			`

ip: 127.0.0.1

`,
			`

ip: 127.0.0.1

ip: ::1

`,
			`

ip: ::1

`,
		},
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", "parent")
			tr.store.Set("/test/data/def", "child")
		},
	},
	templateTest{
		desc: "base64Encode test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/",
]
`,
		tmpl: `
{{$data := base64Encode (getv "/test/data") }}
key: {{$data}}
`,
		expected: `

key: VmFsdWU=
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", `Value`)
		},
	},
	templateTest{
		desc: "base64Decode test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data/",
]
`,
		tmpl: `
{{$data := base64Decode (getv "/test/data") }}
key: {{$data}}
`,
		expected: `

key: Value
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data", `VmFsdWU=`)
		},
	}, templateTest{
		desc: "seq test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
`,
		tmpl: `
{{ seq 1 3 }}
`,
		expected: `
[1 2 3]
`,
		updateStore: func(tr *TemplateResource) {},
	}, templateTest{
		desc: "atoi test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/count/",
]
`,
		tmpl: `
{{ seq 1 (atoi (getv "/test/count")) }}
`,
		expected: `
[1 2 3]
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/count", "3")
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
	switch tt.expected.(type) {
	case string:
		if string(actual) != tt.expected.(string) {
			t.Errorf(fmt.Sprintf("%v: invalid StageFile. Expected %v, actual %v", tt.desc, tt.expected, string(actual)))
		}
	case []string:
		for _, expected := range tt.expected.([]string) {
			if string(actual) == expected {
				break
			}
		}
		t.Errorf(fmt.Sprintf("%v: invalid StageFile. Possible expected values %v, actual %v", tt.desc, tt.expected, string(actual)))
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
