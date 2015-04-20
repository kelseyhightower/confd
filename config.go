package confd

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
	backend           string
	clientCaKeys      string
	clientCert        string
	clientKey         string
	confdir           string
	Cfg               Config // holds the global confd config.
	interval          int
	keepStageFile     bool
	logLevel          string
	nodes             Nodes
	noop              bool
	Onetime           bool
	prefix            string
	PrintVersion      bool
	scheme            string
	srvDomain         string
	TemplateConfig    template.Config
	BackendsConfig    backends.Config
	Watch             bool
)

// A Config structure is used to configure confd.
type Config struct {
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
	LogLevel     string   `toml:"log-level"`
	Watch        bool     `toml:"watch"`
}

func init() {
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
	flag.BoolVar(&Onetime, "onetime", false, "run once and exit")
	flag.StringVar(&prefix, "prefix", "/", "key path prefix")
	flag.BoolVar(&PrintVersion, "version", false, "print version and exit")
	flag.StringVar(&scheme, "scheme", "http", "the backend URI scheme (http or https)")
	flag.StringVar(&srvDomain, "srv-domain", "", "the name of the resource record")
	flag.BoolVar(&Watch, "watch", false, "enable watch support")
}

// InitConfig initializes the confd configuration by first setting defaults,
// then overriding settings from the confd config file, then overriding
// settings from environment variables, and finally overriding
// settings from flags set on the command line.
// It returns an error if any.
func InitConfig() error {
	if configFile == "" {
		if _, err := os.Stat(defaultConfigFile); !os.IsNotExist(err) {
			configFile = defaultConfigFile
		}
	}
	// Set defaults.
	Cfg = Config{
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
		_, err = toml.Decode(string(configBytes), &Cfg)
		if err != nil {
			return err
		}
	}

	// Update config from environment variables.
	processEnv()

	// Update config from commandline flags.
	processFlags()

	if Cfg.LogLevel != "" {
		log.SetLevel(Cfg.LogLevel)
	}

	// Update BackendNodes from SRV records.
	if Cfg.Backend != "env" && Cfg.SRVDomain != "" {
		log.Info("SRV domain set to " + Cfg.SRVDomain)
		srvNodes, err := getBackendNodesFromSRV(Cfg.Backend, Cfg.SRVDomain, Cfg.Scheme)
		if err != nil {
			return errors.New("Cannot get nodes from SRV records " + err.Error())
		}
		Cfg.BackendNodes = srvNodes
	}
	if len(Cfg.BackendNodes) == 0 {
		switch Cfg.Backend {
		case "consul":
			Cfg.BackendNodes = []string{"127.0.0.1:8500"}
		case "etcd":
			peerstr := os.Getenv("ETCDCTL_PEERS")
			if len(peerstr) > 0 {
				Cfg.BackendNodes = strings.Split(peerstr, ",")
			} else {
				Cfg.BackendNodes = []string{"http://127.0.0.1:4001"}
			}
		case "redis":
			Cfg.BackendNodes = []string{"127.0.0.1:6379"}
		}
	}
	// Initialize the storage client
	log.Info("Backend set to " + Cfg.Backend)

	if Cfg.Watch {
		unsupportedBackends := map[string]bool{
			"zookeeper": true,
			"redis":     true,
		}

		if unsupportedBackends[Cfg.Backend] {
			log.Info(fmt.Sprintf("Watch is not supported for backend %s. Exiting...", Cfg.Backend))
			os.Exit(1)
		}
	}

	BackendsConfig = backends.Config{
		Backend:      Cfg.Backend,
		ClientCaKeys: Cfg.ClientCaKeys,
		ClientCert:   Cfg.ClientCert,
		ClientKey:    Cfg.ClientKey,
		BackendNodes: Cfg.BackendNodes,
		Scheme:       Cfg.Scheme,
	}
	// Template configuration.
	TemplateConfig = template.Config{
		ConfDir:       Cfg.ConfDir,
		ConfigDir:     filepath.Join(Cfg.ConfDir, "conf.d"),
		KeepStageFile: keepStageFile,
		Noop:          Cfg.Noop,
		Prefix:        Cfg.Prefix,
		TemplateDir:   filepath.Join(Cfg.ConfDir, "templates"),
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
		Cfg.ClientCaKeys = cakeys
	}

	cert := os.Getenv("CONFD_CLIENT_CERT")
	if len(cert) > 0 {
		Cfg.ClientCert = cert
	}

	key := os.Getenv("CONFD_CLIENT_KEY")
	if len(key) > 0 {
		Cfg.ClientKey = key
	}
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "backend":
		Cfg.Backend = backend
	case "client-cert":
		Cfg.ClientCert = clientCert
	case "client-key":
		Cfg.ClientKey = clientKey
	case "client-ca-keys":
		Cfg.ClientCaKeys = clientCaKeys
	case "confdir":
		Cfg.ConfDir = confdir
	case "node":
		Cfg.BackendNodes = nodes
	case "interval":
		Cfg.Interval = interval
	case "noop":
		Cfg.Noop = noop
	case "prefix":
		Cfg.Prefix = prefix
	case "scheme":
		Cfg.Scheme = scheme
	case "srv-domain":
		Cfg.SRVDomain = srvDomain
	case "log-level":
		Cfg.LogLevel = logLevel
	case "watch":
		Cfg.Watch = Watch
	}
}
