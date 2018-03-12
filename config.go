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
	authType          string
	backend           string
	basicAuth         bool
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
	secretKeyring     string
	scheme            string
	srvDomain         string
	srvRecord         string
	syncOnly          bool
	table             string
	separator         string
	templateConfig    template.Config
	backendsConfig    backends.Config
	username          string
	password          string
	watch             bool
	appID             string
	userID            string
	roleID            string
	secretID          string
	yamlFile          Nodes
	filter            string
)

// A Config structure is used to configure confd.
type Config struct {
	AuthToken     string   `toml:"auth_token"`
	AuthType      string   `toml:"auth_type"`
	Backend       string   `toml:"backend"`
	BasicAuth     bool     `toml:"basic_auth"`
	BackendNodes  []string `toml:"nodes"`
	ClientCaKeys  string   `toml:"client_cakeys"`
	ClientCert    string   `toml:"client_cert"`
	ClientKey     string   `toml:"client_key"`
	ConfDir       string   `toml:"confdir"`
	Interval      int      `toml:"interval"`
	SecretKeyring string   `toml:"secret_keyring"`
	Noop          bool     `toml:"noop"`
	Password      string   `toml:"password"`
	Prefix        string   `toml:"prefix"`
	SRVDomain     string   `toml:"srv_domain"`
	SRVRecord     string   `toml:"srv_record"`
	Scheme        string   `toml:"scheme"`
	SyncOnly      bool     `toml:"sync-only"`
	Table         string   `toml:"table"`
	Separator     string   `toml:"separator"`
	Username      string   `toml:"username"`
	LogLevel      string   `toml:"log-level"`
	Watch         bool     `toml:"watch"`
	AppID         string   `toml:"app_id"`
	UserID        string   `toml:"user_id"`
	RoleID        string   `toml:"role_id"`
	SecretID      string   `toml:"secret_id"`
	YAMLFile      []string `toml:"file"`
	Filter        string   `toml:"filter"`
}

func init() {
	flag.StringVar(&authToken, "auth-token", "", "Auth bearer token to use")
	flag.StringVar(&backend, "backend", "etcd", "backend to use")
	flag.BoolVar(&basicAuth, "basic-auth", false, "Use Basic Auth to authenticate (only used with -backend=consul and -backend=etcd)")
	flag.StringVar(&clientCaKeys, "client-ca-keys", "", "client ca keys")
	flag.StringVar(&clientCert, "client-cert", "", "the client cert")
	flag.StringVar(&clientKey, "client-key", "", "the client key")
	flag.StringVar(&confdir, "confdir", "/etc/confd", "confd conf directory")
	flag.StringVar(&configFile, "config-file", "", "the confd config file")
	flag.Var(&yamlFile, "file", "the YAML file to watch for changes")
	flag.StringVar(&filter, "filter", "*", "files filter (only used with -backend=file)")
	flag.IntVar(&interval, "interval", 600, "backend polling interval")
	flag.BoolVar(&keepStageFile, "keep-stage-file", false, "keep staged files")
	flag.StringVar(&logLevel, "log-level", "", "level which confd should log messages")
	flag.Var(&nodes, "node", "list of backend nodes")
	flag.BoolVar(&noop, "noop", false, "only show pending changes")
	flag.BoolVar(&onetime, "onetime", false, "run once and exit")
	flag.StringVar(&prefix, "prefix", "", "key path prefix")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&scheme, "scheme", "http", "the backend URI scheme for nodes retrieved from DNS SRV records (http or https)")
	flag.StringVar(&secretKeyring, "secret-keyring", "", "path to armored PGP secret keyring (for use with crypt functions)")
	flag.StringVar(&srvDomain, "srv-domain", "", "the name of the resource record")
	flag.StringVar(&srvRecord, "srv-record", "", "the SRV record to search for backends nodes. Example: _etcd-client._tcp.example.com")
	flag.BoolVar(&syncOnly, "sync-only", false, "sync without check_cmd and reload_cmd")
	flag.StringVar(&authType, "auth-type", "", "Vault auth backend type to use (only used with -backend=vault)")
	flag.StringVar(&appID, "app-id", "", "Vault app-id to use with the app-id backend (only used with -backend=vault and auth-type=app-id)")
	flag.StringVar(&userID, "user-id", "", "Vault user-id to use with the app-id backend (only used with -backend=value and auth-type=app-id)")
	flag.StringVar(&roleID, "role-id", "", "Vault role-id to use with the AppRole, Kubernetes backends (only used with -backend=vault and either auth-type=app-role or auth-type=kubernetes)")
	flag.StringVar(&secretID, "secret-id", "", "Vault secret-id to use with the AppRole backend (only used with -backend=vault and auth-type=app-role)")
	flag.StringVar(&table, "table", "", "the name of the DynamoDB table (only used with -backend=dynamodb)")
	flag.StringVar(&separator, "separator", "", "the separator to replace '/' with when looking up keys in the backend, prefixed '/' will also be removed (only used with -backend=redis)")
	flag.StringVar(&username, "username", "", "the username to authenticate as (only used with vault and etcd backends)")
	flag.StringVar(&password, "password", "", "the password to authenticate with (only used with vault and etcd backends)")
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
		Prefix:   "",
		Scheme:   "http",
		Filter:   "*",
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
	var pgpPrivateKey []byte
	if config.SecretKeyring != "" {
		kr, err := os.Open(config.SecretKeyring)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer kr.Close()
		pgpPrivateKey, err = ioutil.ReadAll(kr)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if config.LogLevel != "" {
		log.SetLevel(config.LogLevel)
	}

	if config.SRVDomain != "" && config.SRVRecord == "" {
		config.SRVRecord = fmt.Sprintf("_%s._tcp.%s.", config.Backend, config.SRVDomain)
	}

	// Update BackendNodes from SRV records.
	if config.Backend != "env" && config.SRVRecord != "" {
		log.Info("SRV record set to " + config.SRVRecord)
		srvNodes, err := getBackendNodesFromSRV(config.SRVRecord)
		if err != nil {
			return errors.New("Cannot get nodes from SRV records " + err.Error())
		}

		switch config.Backend {
		case "etcd":
			vsm := make([]string, len(srvNodes))
			for i, v := range srvNodes {
				vsm[i] = config.Scheme + "://" + v
			}
			srvNodes = vsm
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
		case "etcdv3":
			config.BackendNodes = []string{"127.0.0.1:2379"}
		case "redis":
			config.BackendNodes = []string{"127.0.0.1:6379"}
		case "vault":
			config.BackendNodes = []string{"http://127.0.0.1:8200"}
		case "zookeeper":
			config.BackendNodes = []string{"127.0.0.1:2181"}
		}
	}
	// Initialize the storage client
	log.Info("Backend set to " + config.Backend)

	if config.Watch {
		unsupportedBackends := map[string]bool{
			"dynamodb": true,
			"ssm":      true,
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
		AuthType:     config.AuthType,
		Backend:      config.Backend,
		BasicAuth:    config.BasicAuth,
		ClientCaKeys: config.ClientCaKeys,
		ClientCert:   config.ClientCert,
		ClientKey:    config.ClientKey,
		BackendNodes: config.BackendNodes,
		Password:     config.Password,
		Scheme:       config.Scheme,
		Table:        config.Table,
		Separator:    config.Separator,
		Username:     config.Username,
		AppID:        config.AppID,
		UserID:       config.UserID,
		RoleID:       config.RoleID,
		SecretID:     config.SecretID,
		YAMLFile:     config.YAMLFile,
		Filter:       config.Filter,
	}
	// Template configuration.
	templateConfig = template.Config{
		ConfDir:       config.ConfDir,
		ConfigDir:     filepath.Join(config.ConfDir, "conf.d"),
		KeepStageFile: keepStageFile,
		Noop:          config.Noop,
		Prefix:        config.Prefix,
		SyncOnly:      config.SyncOnly,
		TemplateDir:   filepath.Join(config.ConfDir, "templates"),
		PGPPrivateKey: pgpPrivateKey,
	}
	return nil
}

func getBackendNodesFromSRV(record string) ([]string, error) {
	nodes := make([]string, 0)

	// Ignore the CNAME as we don't need it.
	_, addrs, err := net.LookupSRV("", "", record)
	if err != nil {
		return nodes, err
	}
	for _, srv := range addrs {
		host := strings.TrimRight(srv.Target, ".")
		port := strconv.FormatUint(uint64(srv.Port), 10)
		nodes = append(nodes, net.JoinHostPort(host, port))
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
	case "auth-type":
		config.AuthType = authType
	case "backend":
		config.Backend = backend
	case "basic-auth":
		config.BasicAuth = basicAuth
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
	case "password":
		config.Password = password
	case "prefix":
		config.Prefix = prefix
	case "scheme":
		config.Scheme = scheme
	case "secret-keyring":
		config.SecretKeyring = secretKeyring
	case "srv-domain":
		config.SRVDomain = srvDomain
	case "srv-record":
		config.SRVRecord = srvRecord
	case "sync-only":
		config.SyncOnly = syncOnly
	case "table":
		config.Table = table
	case "separator":
		config.Separator = separator
	case "username":
		config.Username = username
	case "log-level":
		config.LogLevel = logLevel
	case "watch":
		config.Watch = watch
	case "app-id":
		config.AppID = appID
	case "user-id":
		config.UserID = userID
	case "role-id":
		config.RoleID = roleID
	case "secret-id":
		config.SecretID = secretID
	case "file":
		config.YAMLFile = yamlFile
	case "filter":
		config.Filter = filter
	}
}
