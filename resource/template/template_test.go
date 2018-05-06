package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kelseyhightower/confd/builtin/databases/env"
	"github.com/kelseyhightower/confd/log"
	"github.com/xordataexchange/crypt/encoding/secconf"
)

const (
	tomlFilePath = "test/confd/config.toml"
	tmplFilePath = "test/templates/test.conf.tmpl"
	cryptPubKey  = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBFTw/PwBCACepvMIcaRYoG19n/e4b/kDNfTNLSseXdiWLVHHP1T30Si/bcsN
57oXWILl/KjM5D8eElAhiQZpmWzN3UJfHxI6q6GO/LxiKS9V2WlILNLSR/IQfr89
xtC6a0s1Fupaq5f8GLpUMYW7hsItkKwAqpqHEQ3pnHNme3lCYAR9FaSzpUjqsfwX
zMqRF4x2T1ZqUvQSC1uPtQlY9vnOmktcvtjha0YvYye7C2++AzzNGT8DjNRzkex3
8eWob1Sd1d9ZC0Ok4IlozuDYHTihbCPFomLfl86V1lKsoh1RaJ4KP/PmRQ7jMr8E
939adfl6Ii9QXQ8bP3DFjW53faIuEl9UiJcNABEBAAG0LWFwcCAoYXBwIGNvbmZp
Z3VyYXRpb24ga2V5KSA8YXBwQGV4YW1wbGUuY29tPokBNwQTAQoAIQUCVPD8/AIb
AwULCQgHAwUVCgkICwUWAgMBAAIeAQIXgAAKCRDIRcPjFDdNtyYLCACO7fSK3U/t
37W9GR8+Mpre9BUiTvgEcZsLzYjkxffqAKl6VfE0mZqHqxxEIW+EYwt3APMZef2Q
v3MruGnOLI0X3qW7h0v3c9XVijBUb3U7n7v120gqjPESVfJEsOT2al8Zw1AZwG9y
ty24xCKob7Mv/3QeIdq9xHUB5iMnlAPixVNJFWL6kgWaAq9y+9MmzNy0WOmHDZlG
SjylyLLEwKjChDDMBsk7RsvNO9l6iz4i1ldz7zxpEs3TiPOM6LFWanA2L8QfOMZl
Z+M9prs50ka2h8neVtPSZtR/mtW/rTHTn5wBEfqGtpRdL0hl9D5fUnpdYw3/cFYB
/OnH6lH5yyBWuQENBFTw/PwBCADk032dt6M9vlZJ+KeBPYZOe+mdAzTiyz+8AFrj
1DI63HC9olXrhmn9S967ilXS/LiYbvrvxmbWSI9qe+rW8ZevBBwpAPhFTsTjAu1E
TuaX8CKYFqY0i4RROwhRLc9+WpY9KdaWZ360LTSO4xb06DobWKUUiPsj/LikDnAR
tnpRLqS40LZNVahSE38Mv+y1ouM3mmSyqalOH1jee6v6CRZ0AM7swmMT8rcsEMRc
kNdFA/qa7FiXRpOERLWauIqx5mCvzZQc7QZnPmGoQsjoCbWl4yKAyoV9GlPwTlAQ
SsFismuePhJAzh1ygHRtpzCNegbZVC6U0PALNs5Z4TZOrSNtABEBAAGJAR8EGAEK
AAkFAlTw/PwCGwwACgkQyEXD4xQ3TbcMRwf/SAJO94jkyH4hh1JqW9lDPZkYhT/1
GLGikREchJSHu+cwesIS7rBwmyUZzabMtqcHwHL5QTsRog+jWwZpZfaD3jTCtRhj
4l/qv94Wy6fUgeu7aFW0AMG8jmZOrSyxl9QrMFPSur8g/yB8i7axbjRJkLwJxNI+
PZMG4yQvvPxFFfeD6RCu4rv8qiHHD8+fh8diksvcaTmx47sZnWEbV3gc9Uvh3bhL
7KPmXrvWQS1PyF3bkaoSEtJtKlqZtebY60zMFKiMmgsW57N2zV5zsjyQmC2rZgBD
7cDGsMeUaF1U6tOzrITSIT1gNnZpojHSSPxddldCee0ACFCQBV/aBrvdtg==
=jz0w
-----END PGP PUBLIC KEY BLOCK-----
`
	cryptPrivKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----

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
tg==
=/iup
-----END PGP PRIVATE KEY BLOCK-----
`
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
		desc: "base, cget test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/crypt-test/key",
]
`,
		tmpl: `
{{with cget "/crypt-test/key"}}
key: {{base .Key}}
val: {{.Value}}
{{end}}
`,
		expected: `

key: key
val: abc

`,
		updateStore: func(tr *TemplateResource) {
			b, err := secconf.Encode([]byte("abc"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-test/key", string(b))
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
		desc: "cgets test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/crypt-test/user",
    "/crypt-test/pass",
    "/crypt-nada/url",
]
`,
		tmpl: `
{{range cgets "/crypt-test/*"}}
key: {{.Key}}
val: {{.Value}}
{{end}}
`,
		expected: `

key: /crypt-test/pass
val: abc

key: /crypt-test/user
val: mary

`,
		updateStore: func(tr *TemplateResource) {
			b, err := secconf.Encode([]byte("mary"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-test/user", string(b))

			b, err = secconf.Encode([]byte("abc"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-test/pass", string(b))

			b, err = secconf.Encode([]byte("url"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-nada/url", string(b))
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
		desc: "cgetv test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/crypt-test/url",
    "/crypt-test/user",
]
`,
		tmpl: `
url = {{cgetv "/crypt-test/url"}}
user = {{cgetv "/crypt-test/user"}}
`,
		expected: `
url = http://www.abc.com
user = bob
`,
		updateStore: func(tr *TemplateResource) {
			b, err := secconf.Encode([]byte("http://www.abc.com"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-test/url", string(b))
			b, err = secconf.Encode([]byte("bob"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-test/user", string(b))
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
		desc: "cgetvs test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/crypt-test/user",
    "/crypt-test/pass",
    "/crypt-nada/url",
]
`,
		tmpl: `
{{range cgetvs "/crypt-test/*"}}
val: {{.}}
{{end}}
`,
		expected: `

val: mary

`,
		updateStore: func(tr *TemplateResource) {
			b, err := secconf.Encode([]byte("mary"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-test/user", string(b))

			// b, err = secconf.Encode([]byte("abc"), bytes.NewBuffer([]byte(cryptPubKey)))
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// tr.store.Set("/crypt-test/pass", string(b))

			b, err = secconf.Encode([]byte("url"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err.Error())
			}
			tr.store.Set("/crypt-nada/url", string(b))
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
	config := Config{
		TemplateDir:   "./test/templates",
		PGPPrivateKey: []byte(cryptPrivKey),
		Database:      &env.Client{},
	}

	tr, err := NewTemplateResource(tomlFilePath, config)
	if err != nil {
		return nil, err
	}
	tr.Dest = "./test/tmp/test.conf"
	tr.FileMode = 0666
	return tr, nil
}
