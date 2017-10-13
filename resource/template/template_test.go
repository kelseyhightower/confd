package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/kelseyhightower/memkv"

	"github.com/kelseyhightower/confd/backends"
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

lQOYBFTw/PwBCACepvMIcaRYoG19n/e4b/kDNfTNLSseXdiWLVHHP1T30Si/bcsN
57oXWILl/KjM5D8eElAhiQZpmWzN3UJfHxI6q6GO/LxiKS9V2WlILNLSR/IQfr89
xtC6a0s1Fupaq5f8GLpUMYW7hsItkKwAqpqHEQ3pnHNme3lCYAR9FaSzpUjqsfwX
zMqRF4x2T1ZqUvQSC1uPtQlY9vnOmktcvtjha0YvYye7C2++AzzNGT8DjNRzkex3
8eWob1Sd1d9ZC0Ok4IlozuDYHTihbCPFomLfl86V1lKsoh1RaJ4KP/PmRQ7jMr8E
939adfl6Ii9QXQ8bP3DFjW53faIuEl9UiJcNABEBAAEAB/wN4Gvi+e+mSdfx1D9a
XD7jS0GffZsnI425avrbatxvdZWzErMfQvy5pIYEgEIyc6dapb7pA/9x1pfX9Mmk
oMbbJ15w74W5r0EC6QqGo9cHyf+v9iobiOuCVrakDN5QMnCPfgk0KoW4Nov+6MfG
oiV0eWcmXwcP+G5NgjD6UN2QUdomZ6zDdt9CSZfkM3o0dvYoVTtUuUFmcMnnSsgN
Dyv4zQpMj0PfTLF06GqPqnU05qYow3Y77XVCOSHkslj4/pJgZb9dkmKFCiusnbFX
lhanxZIjuTtHrT8O/hFl870aDWaDAnrKQw4Un4JRMnDQAdaNSkgkQ5kADIjiYWiG
9ALRBADDRi0Z8AvsBazmnF7o9HsVicOfVrRtaiJqTFShIBnOKq16aoU6AutFqAzy
YmfNrwGe8exhrcneW6bi1EA1CheRX1yAK1zEr307y7G+q2S+lNSHwJRm1ddpE02U
gG1fK/7X9W/w37G4SYmGaihOKlBbDV5mjKLwawtOcIsn0m+QHQQAz/1JsNVttAKn
NZ6ozxa7sdwOLxyKCEePfYTDymvXBpWDnHhUImoogRm69jf1f5r0wmC8/+zy4u2B
r/M4DRo9jtXlDB4ShfA2HMxgu0LACXmACNI59eQ7sYqLyQmwJWDVCHnCDhOBRW8E
dOgVY64ICKUamRj2Z5nCWVOa2U4RT7ED/2OdpdyEAA1oD88Sgy2SMPIDy6OSKuL/
j9z6jNZZ0yJFQ16hf4gEJSilmiK5Q+HdY6EjbZOLHLAXPisRUOnwiVjII5+NNzda
RT9ba31kB1H1ZPpf95653NApdsZw5LsZMFRXobgr6KspEGd2ol8HCuTSNtnL+dQC
95/C5GQ1F2xiPWq0LWFwcCAoYXBwIGNvbmZpZ3VyYXRpb24ga2V5KSA8YXBwQGV4
YW1wbGUuY29tPokBNwQTAQoAIQUCVPD8/AIbAwULCQgHAwUVCgkICwUWAgMBAAIe
AQIXgAAKCRDIRcPjFDdNtyYLCACO7fSK3U/t37W9GR8+Mpre9BUiTvgEcZsLzYjk
xffqAKl6VfE0mZqHqxxEIW+EYwt3APMZef2Qv3MruGnOLI0X3qW7h0v3c9XVijBU
b3U7n7v120gqjPESVfJEsOT2al8Zw1AZwG9yty24xCKob7Mv/3QeIdq9xHUB5iMn
lAPixVNJFWL6kgWaAq9y+9MmzNy0WOmHDZlGSjylyLLEwKjChDDMBsk7RsvNO9l6
iz4i1ldz7zxpEs3TiPOM6LFWanA2L8QfOMZlZ+M9prs50ka2h8neVtPSZtR/mtW/
rTHTn5wBEfqGtpRdL0hl9D5fUnpdYw3/cFYB/OnH6lH5yyBWnQOYBFTw/PwBCADk
032dt6M9vlZJ+KeBPYZOe+mdAzTiyz+8AFrj1DI63HC9olXrhmn9S967ilXS/LiY
bvrvxmbWSI9qe+rW8ZevBBwpAPhFTsTjAu1ETuaX8CKYFqY0i4RROwhRLc9+WpY9
KdaWZ360LTSO4xb06DobWKUUiPsj/LikDnARtnpRLqS40LZNVahSE38Mv+y1ouM3
mmSyqalOH1jee6v6CRZ0AM7swmMT8rcsEMRckNdFA/qa7FiXRpOERLWauIqx5mCv
zZQc7QZnPmGoQsjoCbWl4yKAyoV9GlPwTlAQSsFismuePhJAzh1ygHRtpzCNegbZ
VC6U0PALNs5Z4TZOrSNtABEBAAEAB/0ZEEU2Yj03b8KvJX+LZhFw7/JIUr1B3iCk
AfydRHUgAgc4mNsFCjBPyzX2oBIDIyVA+l7ypm5F0vrKdLo7Qt5qEBmEMD6sBhMG
CSBd7AUmkpSSHgukaRJcJ2AjgXtfC/hgHgC1cVlCePUZytaNNWaFRPIH9sezSw9v
q26BB4C2+sooCC1YLPg8h4YDPdbkoXXRb+0B0XwX9WuEX3WQ9xQtdru3a4KuHtTB
fpWlMzsEvW5io9Qr7Lb3OnSUdyBV3G80X8J3OgX6sM6UhpX7xkoTGIdXpj3vsoTM
6uZmdJ8cakeGtCaoaJmeBsiooCZT/YZ/RJKsPxYyDQ8imzWCgJUtBADlcLXN7uqL
L2XJ+RnU03mnrV9lkeuBFagL74eRkbbf/Et/Qa9PBhWJ0IYK0E4rnAs8EjPYn6YZ
tyN6ur6qDkneedLdZwpACppsiRvPjhMm2yVhM/VhRcF9UV3DxoVaQkF513Co+yOT
NQtdg/XH9tTLmRDaAO04IRJnopvX3FlgiwQA/1CUuwRPl1WXBK9GZKMPkzDqfVkS
lg7smj77W1WfZ+mkdcM75Zk9cpRaRA7r91+Vz50yCKXqVmbnrUvOrE9rZgUuksA8
AgYwe7y3f1wRwxVyAN5pT8xjxOSKXM7mW+E1eQXOccaN8wfpwQWly8YYG6MnVk7O
plDm0RMUJmgf0ucD/A/RbmH9r6Z8euEu2tTo7QGVpHorw7quWRo7VISQuZb3Mnqw
WjY0ErZmCE/aRK2C2XMopLEZESpl14p0YQJhtWFMTRblP4OvHw94SbKpEufhf+rx
GPbD8kYIk3bkp2V/RmFJgcvRvxxxYfYV3G/i4vM2m24ASthIiTXgYaumhlxmM2SJ
AR8EGAEKAAkFAlTw/PwCGwwACgkQyEXD4xQ3TbcMRwf/SAJO94jkyH4hh1JqW9lD
PZkYhT/1GLGikREchJSHu+cwesIS7rBwmyUZzabMtqcHwHL5QTsRog+jWwZpZfaD
3jTCtRhj4l/qv94Wy6fUgeu7aFW0AMG8jmZOrSyxl9QrMFPSur8g/yB8i7axbjRJ
kLwJxNI+PZMG4yQvvPxFFfeD6RCu4rv8qiHHD8+fh8diksvcaTmx47sZnWEbV3gc
9Uvh3bhL7KPmXrvWQS1PyF3bkaoSEtJtKlqZtebY60zMFKiMmgsW57N2zV5zsjyQ
mC2rZgBD7cDGsMeUaF1U6tOzrITSIT1gNnZpojHSSPxddldCee0ACFCQBV/aBrvd
tg==
=/iup
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
				log.Fatal(err)
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
				log.Fatal(err)
			}
			tr.store.Set("/crypt-test/user", string(b))

			b, err = secconf.Encode([]byte("abc"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err)
			}
			tr.store.Set("/crypt-test/pass", string(b))

			b, err = secconf.Encode([]byte("url"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err)
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
				log.Fatal(err)
			}
			tr.store.Set("/crypt-test/url", string(b))
			b, err = secconf.Encode([]byte("bob"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err)
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
				log.Fatal(err)
			}
			tr.store.Set("/crypt-test/user", string(b))

			// b, err = secconf.Encode([]byte("abc"), bytes.NewBuffer([]byte(cryptPubKey)))
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// tr.store.Set("/crypt-test/pass", string(b))

			b, err = secconf.Encode([]byte("url"), bytes.NewBuffer([]byte(cryptPubKey)))
			if err != nil {
				log.Fatal(err)
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
{{range lookupV4IP "localhost"}}
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
		StoreClient:   client, // not used but must be set
		TemplateDir:   "./test/templates",
		PGPPrivateKey: []byte(cryptPrivKey),
	}

	tr, err := NewTemplateResource(tomlFilePath, config)
	if err != nil {
		return nil, err
	}
	tr.Dest = "./test/tmp/test.conf"
	tr.FileMode = 0666
	return tr, nil
}

type tstCompareType int

const (
	tstEq tstCompareType = iota
	tstNe
	tstGt
	tstGe
	tstLt
	tstLe
)

func tstIsEq(tp tstCompareType) bool {
	return tp == tstEq || tp == tstGe || tp == tstLe
}

func tstIsGt(tp tstCompareType) bool {
	return tp == tstGt || tp == tstGe
}

func tstIsLt(tp tstCompareType) bool {
	return tp == tstLt || tp == tstLe
}

var templateExtFuncTests = []templateTest{

	templateTest{
		desc: "add test",
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
val: {{add .Value 1}}
{{end}}
`,
		expected: `

key: key
val: 2

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "1")
		},
	},
	templateTest{
		desc: "sub test",
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
val: {{sub .Value 1}}
{{end}}
`,
		expected: `

key: key
val: 1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "2")
		},
	},
	templateTest{
		desc: "div test",
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
val: {{div .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},
	templateTest{
		desc: "div test2",
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
val: {{div .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 1.5

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3.00")
		},
	},
	templateTest{
		desc: "mul test",
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
val: {{mul .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 6

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},

	templateTest{
		desc: "gt test",
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
val: {{if gt .Value 2}}gt{{end}}
{{end}}
`,
		expected: `

key: key
val: gt

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},
	templateTest{
		desc: "lt test",
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
val: {{if lt .Value 2}}lt{{end}}
{{end}}
`,
		expected: `

key: key
val: lt

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "1")
		},
	},
	templateTest{
		desc: "mod test",
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
val: {{mod .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},
	templateTest{
		desc: "max test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{$nodes := gets "/test/key/*"}}
len: {{len $nodes}}
max: {{max (len $nodes) 3}}
`,
		expected: `

len: 2
max: 3
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key/n1", "v1")
			tr.store.Set("/test/key/n2", "v2")
		},
	},
	templateTest{
		desc: "filter test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/data",
]
`,
		tmpl: `
{{range lsdir "/test/data" | filter "a[123]" }}
value: {{.}}
{{end}}
`,
		expected: `

value: a1

value: a2

value: a3

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/data/a1/v1", "av1")
			tr.store.Set("/test/data/b1/v1", "bv1")
			tr.store.Set("/test/data/a2/v2", "av2")
			tr.store.Set("/test/data/b2/v2", "bv2")
			tr.store.Set("/test/data/a3/v3", "av3")
			tr.store.Set("/test/data/b3/v3", "bv3")
		},
	},
}

func TestFuncsInTemplate(t *testing.T) {
	for _, tt := range templateExtFuncTests {
		ExecuteTestTemplate(tt, t)
	}
}

func TestCompare(t *testing.T) {
	for _, this := range []struct {
		tstCompareType
		funcUnderTest func(a, b interface{}) bool
	}{
		{tstGt, gt},
		{tstLt, lt},
		{tstGe, ge},
		{tstLe, le},
		{tstEq, eq},
		{tstNe, ne},
	} {
		doTestCompare(t, this.tstCompareType, this.funcUnderTest)
	}
}

func toTime(value string) time.Time {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		println(err.Error())
		t = time.Now()
	}
	return t
}

func doTestCompare(t *testing.T, tp tstCompareType, funcUnderTest func(a, b interface{}) bool) {
	for i, this := range []struct {
		left            interface{}
		right           interface{}
		expectIndicator int
	}{
		{5, 8, -1},
		{8, 5, 1},
		{5, 5, 0},
		{int(5), int64(5), 0},
		{int32(5), int(5), 0},
		{int16(4), int(5), -1},
		{uint(15), uint64(15), 0},
		{-2, 1, -1},
		{2, -5, 1},
		{0.0, 1.23, -1},
		{1.1, 1.1, 0},
		{float32(1.0), float64(1.0), 0},
		{1.23, 0.0, 1},
		{"5", "5", 0},
		{"5", 5, 0},
		{5, "5", 0},
		{5.1, "5.1", 0},
		{5.0, "5", 0},
		{5, "5.0", 0},
		{"8", 5, 1},
		{5, "8", -1},
		{"8", "5", 1},
		{"8", "5.1", 1},
		{"8", 5.1, 1},
		{8, "5.1", 1},
		{"5", "0001", 1},
		{"a", "a", 0},
		{"a", "b", -1},
		{"b", "a", 1},
		{[]int{100, 99}, []int{1, 2, 3, 4}, -1},
		{toTime("2015-11-20"), toTime("2015-11-20"), 0},
		{toTime("2015-11-19"), toTime("2015-11-20"), -1},
		{toTime("2015-11-20"), toTime("2015-11-19"), 1},
	} {
		result := funcUnderTest(this.left, this.right)
		success := false

		if this.expectIndicator == 0 {
			if tstIsEq(tp) {
				success = result
			} else {
				success = !result
			}
		}

		if this.expectIndicator < 0 {
			success = result && (tstIsLt(tp) || tp == tstNe)
			success = success || (!result && !tstIsLt(tp))
		}

		if this.expectIndicator > 0 {
			success = result && (tstIsGt(tp) || tp == tstNe)
			success = success || (!result && (!tstIsGt(tp) || tp != tstNe))
		}

		if !success {
			t.Errorf("[%d][%s] %v compared to %v: %t", i, path.Base(runtime.FuncForPC(reflect.ValueOf(funcUnderTest).Pointer()).Name()), this.left, this.right, result)
		}
	}
}

func TestMod(t *testing.T) {
	for i, this := range []struct {
		a      interface{}
		b      interface{}
		expect interface{}
	}{
		{3, 2, int64(1)},
		{"3", 2, int64(1)},
		{3, "2", int64(1)},
		{3, 1, int64(0)},
		{3, 0, false},
		{0, 3, int64(0)},
		{3.1, 2, false},
		{3, 2.1, false},
		{3.1, 2.1, false},
		{int8(3), int8(2), int64(1)},
		{int16(3), int16(2), int64(1)},
		{int32(3), int32(2), int64(1)},
		{int64(3), int64(2), int64(1)},
	} {
		result, err := mod(this.a, this.b)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] modulo didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] modulo got %v but expected %v", i, result, this.expect)
			}
		}
	}
}

func TestMaxAndMin(t *testing.T) {
	for i, this := range []struct {
		a      interface{}
		b      interface{}
		expect interface{}
	}{
		{3, 2, float64(3)},
		{"3", 2, float64(3)},
		{3, "2", float64(3)},
		{3.1, 3, float64(3.1)},
		{3, "a", false},
		{int8(3), int8(2), float64(3)},
		{int16(3), int16(2), float64(3)},
		{int32(3), int32(2), float64(3)},
		{int64(3), int64(2), float64(3)},
		{float64(3.0001), float64(3.00011), float64(3.00011)},
	} {
		result, err := max(this.a, this.b)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] max didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] max got %v but expected %v", i, result, this.expect)
			}
		}

		result, err = min(this.a, this.b)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] min didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] min not expected %v", i, result)
			}
		}
	}
}

func TestTimeUnix(t *testing.T) {
	var sec int64 = 1234567890
	tv := reflect.ValueOf(time.Unix(sec, 0))
	i := 1

	res := toTimeUnix(tv)
	if sec != res {
		t.Errorf("[%d] timeUnix got %v but expected %v", i, res, sec)
	}

	i++
	func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("[%d] timeUnix didn't return an expected error", i)
			}
		}()
		iv := reflect.ValueOf(sec)
		toTimeUnix(iv)
	}(t)
}

func TestDoArithmetic(t *testing.T) {
	for i, this := range []struct {
		a      interface{}
		b      interface{}
		op     rune
		expect interface{}
	}{
		{3, 2, '+', int64(5)},
		{3, 2, '-', int64(1)},
		{3, 2, '*', int64(6)},
		{3, 2, '/', int64(1)},
		{3.0, 2, '+', float64(5)},
		{3.0, 2, '-', float64(1)},
		{3.0, 2, '*', float64(6)},
		{3.0, 2, '/', float64(1.5)},
		{3, 2.0, '+', float64(5)},
		{3, 2.0, '-', float64(1)},
		{3, 2.0, '*', float64(6)},
		{3, 2.0, '/', float64(1.5)},
		{3.0, 2.0, '+', float64(5)},
		{3.0, 2.0, '-', float64(1)},
		{3.0, 2.0, '*', float64(6)},
		{3.0, 2.0, '/', float64(1.5)},
		{uint(3), uint(2), '+', int64(5)},
		{uint(3), uint(2), '-', int64(1)},
		{uint(3), uint(2), '*', int64(6)},
		{uint(3), uint(2), '/', int64(1)},
		{uint(3), 2, '+', int64(5)},
		{uint(3), 2, '-', int64(1)},
		{uint(3), 2, '*', int64(6)},
		{uint(3), 2, '/', int64(1)},
		{3, uint(2), '+', int64(5)},
		{3, uint(2), '-', int64(1)},
		{3, uint(2), '*', int64(6)},
		{3, uint(2), '/', int64(1)},
		{uint(3), -2, '+', int64(1)},
		{uint(3), -2, '-', int64(5)},
		{uint(3), -2, '*', int64(-6)},
		{uint(3), -2, '/', int64(-1)},
		{-3, uint(2), '+', int64(-1)},
		{-3, uint(2), '-', int64(-5)},
		{-3, uint(2), '*', int64(-6)},
		{-3, uint(2), '/', int64(-1)},
		{uint(3), 2.0, '+', float64(5)},
		{uint(3), 2.0, '-', float64(1)},
		{uint(3), 2.0, '*', float64(6)},
		{uint(3), 2.0, '/', float64(1.5)},
		{3.0, uint(2), '+', float64(5)},
		{3.0, uint(2), '-', float64(1)},
		{3.0, uint(2), '*', float64(6)},
		{3.0, uint(2), '/', float64(1.5)},
		{0, 0, '+', 0},
		{0, 0, '-', 0},
		{0, 0, '*', 0},
		{"foo", "bar", '+', false},
		{3, 0, '/', false},
		{3.0, 0, '/', false},
		{3, 0.0, '/', false},
		{uint(3), uint(0), '/', false},
		{3, uint(0), '/', false},
		{-3, uint(0), '/', false},
		{uint(3), 0, '/', false},
		{3.0, uint(0), '/', false},
		{uint(3), 0.0, '/', false},
		{3, "foo", '+', false},
		{3.0, "foo", '+', false},
		{uint(3), "foo", '+', false},
		{"foo", 3, '+', false},
		{"foo", "bar", '-', false},
		{"3", "2", '+', int64(5)},
		{"3", "2", '-', int64(1)},
		{"3", "2", '*', int64(6)},
		{"3", "2", '/', int64(1)},
		{"3.0", "2", '+', float64(5)},
		{"3.0", "2", '-', float64(1)},
		{"3.0", "2", '*', float64(6)},
		{"3.0", "2", '/', float64(1.5)},
		{"3", "2.0", '+', float64(5)},
		{"3", "2.0", '-', float64(1)},
		{"3", "2.0", '*', float64(6)},
		{"3", "2.0", '/', float64(1.5)},
		{"3.0", "2.0", '+', float64(5)},
		{"3.0", "2.0", '-', float64(1)},
		{"3.0", "2.0", '*', float64(6)},
		{"3.0", "2.0", '/', float64(1.5)},
	} {
		result, err := DoArithmetic(this.a, this.b, this.op)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] doArithmetic didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] doArithmetic [%v %s %v ] got %v but expected %v", i, this.a, string(this.op), this.b, result, this.expect)
			}
		}
	}
}

func TestFilter(t *testing.T) {
	for i, this := range []struct {
		input  interface{}
		regex  string
		err    bool
		expect interface{}
	}{
		{[]string{"a1", "b1", "a2", "b2", "a3", "b3"}, "a[123]", false, []interface{}{"a1", "a2", "a3"}},
		{[6]string{"a1", "b1", "a2", "b2", "a3", "b3"}, "a[123]", false, []interface{}{"a1", "a2", "a3"}},
		{[]interface{}{"a1", 1, "a2", 2, "a3", 3}, "a[123]", false, []interface{}{"a1", "a2", "a3"}},
		{[6]interface{}{"a1", 1, "a2", 2, "a3", 3}, "a[123]", false, []interface{}{"a1", "a2", "a3"}},
		{[]interface{}{"a1"}, "a.**", true, nil},
		{"a1", "a.**", true, nil},
		{[]interface{}{
			memkv.KVPair{Key: "k1", Value: "a1"},
			memkv.KVPair{Key: "k2", Value: "a2"},
			memkv.KVPair{Key: "k3", Value: "a3"},
			memkv.KVPair{Key: "k4", Value: "b1"},
		},

			"a[123]", false, []interface{}{
				memkv.KVPair{Key: "k1", Value: "a1"},
				memkv.KVPair{Key: "k2", Value: "a2"},
				memkv.KVPair{Key: "k3", Value: "a3"},
			}},
	} {
		result, err := Filter(this.regex, this.input)
		if this.err {
			if err == nil {
				t.Errorf("[%d] Filter didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] Filter [%v] got %v but expected %v", i, this.input, result, this.expect)
			}
		}
	}
}

func TestToJsonAndYaml(t *testing.T) {
	for i, this := range []struct {
		input      interface{}
		err        bool
		expectJson string
		expectYaml string
	}{
		{[]string{"a1", "b1"}, false, `["a1","b1"]`, "- a1\n- b1\n"},
		{"a1", false, `"a1"`, "a1\n"},
		{1, false, "1", "1\n"},
		{struct{ Name string }{Name: "test"}, false, `{"Name":"test"}`, "name: test\n"},
		{struct {
			Name string
			Addr []string
		}{Name: "test", Addr: []string{"a1", "a2"}}, false, `{"Name":"test","Addr":["a1","a2"]}`,
			`name: test
addr:
- a1
- a2
`},
	} {
		result, err := ToJson(this.input)
		if this.err {
			if err == nil {
				t.Errorf("[%d] ToJson didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expectJson) {
				t.Errorf("[%d] ToJson [%v] got %v but expected %v", i, this.input, result, this.expectJson)
			}
		}

		result, err = ToYaml(this.input)
		if this.err {
			if err == nil {
				t.Errorf("[%d] ToYaml didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if result != this.expectYaml {
				t.Errorf("[%d] ToYaml [%v] got %v but expected %v", i, this.input, result, this.expectYaml)
			}
		}
	}
}
