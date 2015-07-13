package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
)

var (
	configFile        = ""
	defaultConfigFile = "/etc/confd/confd.toml"
	authToken         string
	backend           string
	clientCaKeys      string
	clientCert        string
	clientKey         string
	confdir           string
	config            Config // holds the global confd config.
	interval          int
	keepStageFile     bool
	logLevel          string
	nodes             Nodes
	noop              bool
	onetime           bool
	prefix            string
	printVersion      bool
	scheme            string
	srvDomain         string
	table             string
	templateConfig    template.Config
	backendsConfig    backends.Config
	watch             bool
)

// A Config structure is used to configure confd.
type Config struct {
	AuthToken    string   `toml:"auth_token"`
	Backend      string   `toml:"backend"`
	BackendNodes []string `toml:"nodes"`
	ClientCaKeys string   `toml:"client_cakeys"`
	ClientCert   string   `toml:"client_cert"`
	ClientKey    string   `toml:"client_key"`
	ConfDir      string   `toml:"confdir"`
	Interval     int      `toml:"interval"`
	Noop         bool     `toml:"noop"`
	Prefix       string   `toml:"prefix"`
	SRVDomain    string   `toml:"srv_domain"`
	Scheme       string   `toml:"scheme"`
	Table        string   `toml:"table"`
	LogLevel     string   `toml:"log-level"`
	Watch        bool     `toml:"watch"`
}

func init() {
	flag.StringVar(&authToken, "auth-token", "", "Auth bearer token to use")
	flag.StringVar(&backend, "backend", "etcd", "backend to use")
	flag.StringVar(&clientCaKeys, "client-ca-keys", "", "client ca keys")
	flag.StringVar(&clientCert, "client-cert", "", "the client cert")
	flag.StringVar(&clientKey, "client-key", "", "the client key")
	flag.StringVar(&confdir, "confdir", "/etc/confd", "confd conf directory")
	flag.StringVar(&configFile, "config-file", "", "the confd config file")
	flag.IntVar(&interval, "interval", 600, "backend polling interval")
	flag.BoolVar(&keepStageFile, "keep-stage-file", false, "keep staged files")
	flag.StringVar(&logLevel, "log-level", "", "level which confd should log messages")
	flag.Var(&nodes, "node", "list of backend nodes")
	flag.BoolVar(&noop, "noop", false, "only show pending changes")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
	flag.StringVar(&prefix, "prefix", "/", "key path prefix")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&scheme, "scheme", "http", "the backend URI scheme (http or https)")
	flag.StringVar(&srvDomain, "srv-domain", "", "the name of the resource record")
	flag.StringVar(&table, "table", "", "the name of the DynamoDB table (only used with -backend=dynamodb)")
	flag.BoolVar(&watch, "watch", false, "enable watch support")
}

// initConfig initializes the confd configuration by first setting defaults,
// then overriding settings from the confd config file, then overriding
// settings from environment variables, and finally overriding
// settings from flags set on the command line.
// It returns an error if any.
func initConfig() error {
	if configFile == "" {
		if _, err := os.Stat(defaultConfigFile); !os.IsNotExist(err) {
			configFile = defaultConfigFile
		}
	}
	// Set defaults.
	config = Config{
		Backend:  "etcd",
		ConfDir:  "/etc/confd",
		Interval: 600,
		Prefix:   "/",
		Scheme:   "http",
	}
	// Update config from the TOML configuration file.
	if configFile == "" {
		log.Debug("Skipping confd config file.")
	} else {
		log.Debug("Loading " + configFile)
		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}

	// Update config from environment variables.
	processEnv()

	// Update config from commandline flags.
	processFlags()

	if config.LogLevel != "" {
		log.SetLevel(config.LogLevel)
	}

	// Update BackendNodes from SRV records.
	if config.Backend != "env" && config.SRVDomain != "" {
		log.Info("SRV domain set to " + config.SRVDomain)
		srvNodes, err := getBackendNodesFromSRV(config.Backend, config.SRVDomain, config.Scheme)
		if err != nil {
			return errors.New("Cannot get nodes from SRV records " + err.Error())
		}
		config.BackendNodes = srvNodes
	}
	if len(config.BackendNodes) == 0 {
		switch config.Backend {
		case "consul":
			config.BackendNodes = []string{"127.0.0.1:8500"}
		case "etcd":
			peerstr := os.Getenv("ETCDCTL_PEERS")
			if len(peerstr) > 0 {
				config.BackendNodes = strings.Split(peerstr, ",")
			} else {
				config.BackendNodes = []string{"http://127.0.0.1:4001"}
			}
		case "redis":
			config.BackendNodes = []string{"127.0.0.1:6379"}
		case "zookeeper":
			config.BackendNodes = []string{"127.0.0.1:2181"}
		}
	}
	// Initialize the storage client
	log.Info("Backend set to " + config.Backend)

	if config.Watch {
		unsupportedBackends := map[string]bool{
			"zookeeper": true,
			"redis":     true,
			"dynamodb":  true,
		}

		if unsupportedBackends[config.Backend] {
			log.Info(fmt.Sprintf("Watch is not supported for backend %s. Exiting...", config.Backend))
			os.Exit(1)
		}
	}

	if config.Backend == "dynamodb" && config.Table == "" {
		return errors.New("No DynamoDB table configured")
	}

	backendsConfig = backends.Config{
		AuthToken:    config.AuthToken,
		Backend:      config.Backend,
		ClientCaKeys: config.ClientCaKeys,
		ClientCert:   config.ClientCert,
		ClientKey:    config.ClientKey,
		BackendNodes: config.BackendNodes,
		Scheme:       config.Scheme,
		Table:        config.Table,
	}
	// Template configuration.
	templateConfig = template.Config{
		ConfDir:       config.ConfDir,
		ConfigDir:     filepath.Join(config.ConfDir, "conf.d"),
		KeepStageFile: keepStageFile,
		Noop:          config.Noop,
		Prefix:        config.Prefix,
		TemplateDir:   filepath.Join(config.ConfDir, "templates"),
	}
	return nil
}

func getBackendNodesFromSRV(backend, domain, scheme string) ([]string, error) {
	nodes := make([]string, 0)
	// Ignore the CNAME as we don't need it.
	_, addrs, err := net.LookupSRV(backend, "tcp", domain)
	if err != nil {
		return nodes, err
	}
	for _, srv := range addrs {
		host := strings.TrimRight(srv.Target, ".")
		port := strconv.FormatUint(uint64(srv.Port), 10)
		nodes = append(nodes, fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(host, port)))
	}
	return nodes, nil
}

// processFlags iterates through each flag set on the command line and
// overrides corresponding configuration settings.
func processFlags() {
	flag.Visit(setConfigFromFlag)
}

func processEnv() {
	cakeys := os.Getenv("CONFD_CLIENT_CAKEYS")
	if len(cakeys) > 0 {
		config.ClientCaKeys = cakeys
	}

	cert := os.Getenv("CONFD_CLIENT_CERT")
	if len(cert) > 0 {
		config.ClientCert = cert
	}

	key := os.Getenv("CONFD_CLIENT_KEY")
	if len(key) > 0 {
		config.ClientKey = key
	}
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "auth-token":
		config.AuthToken = authToken
	case "backend":
		config.Backend = backend
	case "client-cert":
		config.ClientCert = clientCert
	case "client-key":
		config.ClientKey = clientKey
	case "client-ca-keys":
		config.ClientCaKeys = clientCaKeys
	case "confdir":
		config.ConfDir = confdir
	case "node":
		config.BackendNodes = nodes
	case "interval":
		config.Interval = interval
	case "noop":
		config.Noop = noop
	case "prefix":
		config.Prefix = prefix
	case "scheme":
		config.Scheme = scheme
	case "srv-domain":
		config.SRVDomain = srvDomain
	case "table":
		config.Table = table
	case "log-level":
		config.LogLevel = logLevel
	case "watch":
		config.Watch = watch
	}
}
