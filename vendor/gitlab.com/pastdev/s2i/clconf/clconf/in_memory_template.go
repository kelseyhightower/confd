package clconf

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/kelseyhightower/memkv"
	clconftemplate "gitlab.com/pastdev/s2i/clconf/clconf/template"
)

// TemplateConfig allows for optional configuration.
type TemplateConfig struct {
	Prefix      string
	SecretAgent *SecretAgent
}

// Template is a wrapper for template.Template to include custom template
// functions corresponding to confd functions.
type Template struct {
	config   *TemplateConfig
	store    *memkv.Store
	template *template.Template
}

/////// mapped to confd resource.go ///////
func addCryptFuncs(funcMap map[string]interface{}, sa *SecretAgent) {
	clconftemplate.AddFuncs(funcMap, map[string]interface{}{
		"cget": func(key string) (memkv.KVPair, error) {
			kv, err := funcMap["get"].(func(string) (memkv.KVPair, error))(key)
			if err == nil {
				var decrypted string
				decrypted, err = sa.Decrypt(kv.Value)
				if err == nil {
					kv.Value = decrypted
				}
			}
			return kv, err
		},
		"cgets": func(pattern string) (memkv.KVPairs, error) {
			kvs, err := funcMap["gets"].(func(string) (memkv.KVPairs, error))(pattern)
			if err == nil {
				for i := range kvs {
					decrypted, err := sa.Decrypt(kvs[i].Value)
					if err != nil {
						return memkv.KVPairs(nil), err
					}
					kvs[i].Value = decrypted
				}
			}
			return kvs, err
		},
		"cgetv": func(key string) (string, error) {
			v, err := funcMap["getv"].(func(string, ...string) (string, error))(key)
			if err == nil {
				var decrypted string
				decrypted, err = sa.Decrypt(v)
				if err == nil {
					return decrypted, nil
				}
			}
			return v, err
		},
		"cgetvs": func(pattern string) ([]string, error) {
			vs, err := funcMap["getvs"].(func(string) ([]string, error))(pattern)
			if err == nil {
				for i := range vs {
					decrypted, err := sa.Decrypt(vs[i])
					if err != nil {
						return []string(nil), err
					}
					vs[i] = decrypted
				}
			}
			return vs, err
		},
	})
}

// NewTemplate returns a parsed Template configured with standard functions.
func NewTemplate(name, text string, config *TemplateConfig) (*Template, error) {
	if config == nil {
		config = &TemplateConfig{}
	}

	store := memkv.New()

	funcMap := clconftemplate.NewFuncMap()
	clconftemplate.AddFuncs(funcMap, store.FuncMap)
	if config.SecretAgent != nil {
		addCryptFuncs(funcMap, config.SecretAgent)
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(text)
	if err != nil {
		return nil, fmt.Errorf("Unable to process template %s: %s", name, err)
	}

	return &Template{
		config:   config,
		store:    &store,
		template: tmpl,
	}, nil
}

// NewTemplateFromBase64 decodes base64 then calls NewTemplate with the result.
func NewTemplateFromBase64(name, base64 string, config *TemplateConfig) (*Template, error) {
	contents, err := DecodeBase64Strings(base64)
	if err != nil {
		return nil, err
	}
	return NewTemplate(name, contents[0], config)
}

// NewTemplateFromFile reads file then calls NewTemplate with the result.
func NewTemplateFromFile(name, file string, config *TemplateConfig) (*Template, error) {
	contents, err := ReadFiles(file)
	if err != nil {
		return nil, err
	}
	return NewTemplate(name, contents[0], config)
}

// Execute will process the template text using data and the function map from
// confd.
func (tmpl *Template) Execute(data interface{}) (string, error) {
	tmpl.setVars(data)

	var buf bytes.Buffer
	if err := tmpl.template.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}

/////// mapped to confd resource.go ///////
func (tmpl *Template) setVars(data interface{}) {
	value, _ := GetValue(data, tmpl.config.Prefix)
	tmpl.store.Purge()
	for k, v := range ToKvMap(value) {
		tmpl.store.Set(k, v)
	}
}
