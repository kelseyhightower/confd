package confd

import (
	"github.com/BurntSushi/toml"
	"path/filepath"
)

type Config struct {
	Path     string
	Template Template
}

func FindConfigs(confDir string) ([]string, error) {
	return filepath.Glob(filepath.Join(confDir, "*toml"))
}

func NewConfig(path string) (*Config, error) {
	c := &Config{
		Path: path,
	}
	if _, err := toml.DecodeFile(path, &c.Template); err != nil {
		return c, err
	}
	return c, nil
}

func (c *Config) Process() error {
	if err := c.Template.Process(); err != nil {
		return err
	}
	return nil
}
