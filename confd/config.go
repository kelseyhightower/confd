package confd

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Path      string
	Templates []Template
	TmplDir   string
}

func FindConfigs(configDir string) ([]string, error) {
	return filepath.Glob(filepath.Join(configDir, "*json"))
}

func NewConfig(path string) (*Config, error) {
	var c *Config
	c.Path = path
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(f, &c); err != nil {
		return c, err
	}
	return c, nil
}

func (c *Config) Process() error {
	for _, t := range c.Templates {
		if err := t.Process(); err != nil {
			return err
		}
	}
	return nil
}
