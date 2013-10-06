package config

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
)

func init() {
	flag.Var(&nodes, "n", "list of etcd nodes")
	flag.StringVar(&confdir, "c", "/etc/confd", "confd config directory")
	flag.StringVar(&confFile, "C", "/etc/confd/confd.toml", "confd config file")
	flag.IntVar(&interval, "i", 600, "etcd polling interval")
	flag.StringVar(&prefix, "p", "/", "etcd key path prefix")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
}

var (
	config   Config
	confFile = "/etc/confd/confd.toml"
	nodes    Nodes
	confdir  string
	interval int
	prefix   string
	onetime  bool
)

type Nodes []string

func (n *Nodes) String() string {
	return fmt.Sprintf("%d", *n)
}

func (n *Nodes) Set(node string) error {
	*n = append(*n, node)
	return nil
}

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

func Onetime() bool {
	return onetime
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

func setDefaults() {
	config = Config{
		ConfDir:   "/etc/confd/conf.d",
		Interval:  600,
		Prefix:    "/",
		EtcdNodes: []string{"http://127.0.0.1:4001"},
	}
}

func loadConfFile() error {
	if IsFileExist(confFile) {
		_, err := toml.DecodeFile(confFile, &config)
		if err != nil {
			return err
		}
	}
	return nil
}

func override(f *flag.Flag) {
	switch f.Name {
	case "c":
		config.ConfDir = confdir
	case "i":
		config.Interval = interval
	case "n":
		config.EtcdNodes = nodes
	case "p":
		config.Prefix = prefix
	}
}

func overrideConfig() {
	flag.Visit(override)
}

func SetConfig() error {
	setDefaults()
	if err := loadConfFile(); err != nil {
		return err
	}
	overrideConfig()
	return nil
}

func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
