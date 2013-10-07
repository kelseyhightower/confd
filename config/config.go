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
	Confd confd
}

type confd struct {
	ConfDir   string
	Interval  int
	Prefix    string
	EtcdNodes []string `toml:"etcd_nodes"`
}

func ConfigDir() string {
	return filepath.Join(config.Confd.ConfDir, "conf.d")
}

func EtcdNodes() []string {
	return config.Confd.EtcdNodes
}

func Interval() int {
	return config.Confd.Interval
}

func Onetime() bool {
	return onetime
}

func Prefix() string {
	return config.Confd.Prefix
}

func TemplateDir() string {
	return filepath.Join(config.Confd.ConfDir, "templates")
}

func SetConfFile(path string) {
	confFile = path
}

func setDefaults() {
	config = Config{
		Confd: confd{
			ConfDir:   "/etc/confd/conf.d",
			Interval:  600,
			Prefix:    "/",
			EtcdNodes: []string{"http://127.0.0.1:4001"},
		},
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
		config.Confd.ConfDir = confdir
	case "i":
		config.Confd.Interval = interval
	case "n":
		config.Confd.EtcdNodes = nodes
	case "p":
		config.Confd.Prefix = prefix
	}
}

func overrideConfig() {
	flag.Visit(override)
}

func InitConfig() error {
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
