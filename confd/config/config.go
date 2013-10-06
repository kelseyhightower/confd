package config

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
)

func init() {
	config = Config{
		ConfDir:   "/etc/confd/conf.d",
		Interval:  600,
		Prefix:    "/",
		EtcdNodes: []string{"http://127.0.0.1:4001"},
	}
}

var (
	config   Config
	confFile = "/etc/confd/confd.toml"
)

type Config struct {
	ConfDir   string
	Interval  int
	Prefix    string
	EtcdNodes []string `toml:"etcd_nodes"`
}

func ConfDir() string {
	return config.ConfDir
}

func EtcdNodes() []string {
	return config.EtcdNodes
}

func Interval() int {
	return config.Interval
}

func Prefix() string {
	return config.Prefix
}

func TemplateDir() string {
	return filepath.Join(config.ConfDir, "templates")
}

func SetConfFile(path string) {
	confFile = path
}

func SetConfig() error {
	if IsFileExist(confFile) {
		_, err := toml.DecodeFile(confFile, &config)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
