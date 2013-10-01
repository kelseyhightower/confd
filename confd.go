package main

import (
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/go-ini"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Settings struct {
	ConfigDir  string
	EtcdURL    string
	EtcdPrefix string
}

type Config struct {
	Templates []Template
	Services  map[string]Service
}

type Service struct {
	Name      string
	ReloadCmd string `json:"reload_cmd"`
}

type Template struct {
	Dest    string
	Group   string
	Keys    []string
	Mode    string
	Owner   string
	Service string
	Src     string
}

var settings Settings
var defaultConfig = "/etc/confd/confd.ini"

func main() {
	if err := setConfig(); err != nil {
		log.Fatal(err.Error())
	}
	c := etcd.NewClient()

	// Get config
	configs, err := filepath.Glob(filepath.Join(settings.ConfigDir, "*json"))
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, config := range configs {
		f, err := ioutil.ReadFile(config)
		if err != nil {
			log.Fatal(err.Error())
		}

		var cfg *Config
		err = json.Unmarshal(f, &cfg)
		if err != nil {
			log.Fatal(err.Error())
		}

		// Get the values we care about.
		r := strings.NewReplacer("/", "_")
		for _, tmplConfig := range cfg.Templates {
			m := make(map[string]interface{})
			for _, key := range tmplConfig.Keys {
				values, err := c.Get(filepath.Join(settings.EtcdPrefix, key))
				if err != nil {
					log.Fatal(err.Error())
				}
				for _, v := range values {
					key := strings.TrimPrefix(v.Key, settings.EtcdPrefix)
					new_key := r.Replace(key)
					m[new_key] = v.Value
				}
			}
			tmpl := filepath.Join(settings.ConfigDir, "templates", tmplConfig.Src)
			if isFileExist(tmpl) {
				temp, err := ioutil.TempFile("", "")
				defer os.Remove(temp.Name())
				if err != nil {
					log.Fatal(err.Error())
				}

				data, err := ioutil.ReadFile(tmpl)
				if err != nil {
					log.Fatal(err.Error())
				}
				t := template.Must(template.New("test").Parse(string(data)))
				err = t.Execute(temp, m)
				if err != nil {
					log.Fatal(err.Error())
				}

				// Compare group and owner
				// Compare file mode
				// Compare file contents
				// If anything out of sync copy the temp file to the dest
				// and reload the service.
				os.Rename(temp.Name(), tmplConfig.Dest)
			} else {
				log.Fatal("Missing template: " + tmpl)
			}
		}
	}
}

func setConfig() error {
	// Set defaults
	settings.ConfigDir = "/etc/confd/conf.d"
	settings.EtcdURL = "http://0.0.0.0:4001"
	settings.EtcdPrefix = "/"

	// Override defaults from config file.
	if isFileExist(defaultConfig) {
		s, err := ini.LoadFile(defaultConfig)
		if err != nil {
			return err
		}

		if configDir, ok := s.Get("main", "config_dir"); ok {
			settings.ConfigDir = configDir
		}
		if etcdURL, ok := s.Get("etcd", "url"); ok {
			settings.EtcdURL = etcdURL
		}
		if etcdPrefix, ok := s.Get("etcd", "prefix"); ok {
			settings.EtcdPrefix = etcdPrefix
		}
	}
	return nil
}

func isFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
