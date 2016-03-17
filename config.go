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
	scheme            string
	srvDomain         string
	srvRecord         string
	syncOnly          bool
	table             string
	templateConfig    template.Config
	backendsConfig    backends.Config
	backendsConfig2   backends.Config
	username          string
	password          string
	watch             bool
	appID             string
	userID            string
	// 2nd backend config
	authToken2         string
	authType2          string
	backend2           string
	basicAuth2         bool
	clientCaKeys2      string
	clientCert2        string
	clientKey2         string
	nodes2             Nodes
	prefix2            string
	table2             string
	username2          string
	password2          string
	appID2             string
	userID2            string
)

// A Config structure is used to configure confd.
type Config struct {
	AuthToken    string   `toml:"auth_token"`
	AuthType     string   `toml:"auth_type"`
	Backend      string   `toml:"backend"`
	BasicAuth    bool     `toml:"basic_auth"`
	BackendNodes []string `toml:"nodes"`
	ClientCaKeys string   `toml:"client_cakeys"`
	ClientCert   string   `toml:"client_cert"`
	ClientKey    string   `toml:"client_key"`
	ConfDir      string   `toml:"confdir"`
	Interval     int      `toml:"interval"`
	Noop         bool     `toml:"noop"`
	Password     string   `toml:"password"`
	Prefix       string   `toml:"prefix"`
	SRVDomain    string   `toml:"srv_domain"`
	SRVRecord    string   `toml:"srv_record"`
	Scheme       string   `toml:"scheme"`
	SyncOnly     bool     `toml:"sync-only"`
	Table        string   `toml:"table"`
	Username     string   `toml:"username"`
	LogLevel     string   `toml:"log-level"`
	Watch        bool     `toml:"watch"`
	AppID        string   `toml:"app_id"`
	UserID       string   `toml:"user_id"`
	// 2nd backend config
	AuthToken2    string   `toml:"auth_token2"`
	AuthType2     string   `toml:"auth_type2"`
	Backend2      string   `toml:"backend2"`
	BasicAuth2    bool     `toml:"basic_auth2"`
	BackendNodes2 []string `toml:"nodes2"`
	ClientCaKeys2 string   `toml:"client_cakeys2"`
	ClientCert2   string   `toml:"client_cert2"`
	ClientKey2    string   `toml:"client_key2"`
	Password2     string   `toml:"password2"`
	Prefix2       string   `toml:"prefix2"`
	Table2        string   `toml:"table2"`
	Username2     string   `toml:"username2"`
	AppID2        string   `toml:"app_id2"`
	UserID2       string   `toml:"user_id2"`
}

func init() {
	flag.StringVar(&authToken, "auth-token", "", "Auth bearer token to use")
	flag.StringVar(&backend, "backend", "etcd", "backend to use")
	flag.BoolVar(&basicAuth, "basic-auth", false, "Use Basic Auth to authenticate (only used with -backend=etcd)")
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
	flag.StringVar(&prefix, "prefix", "", "key path prefix")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&scheme, "scheme", "http", "the backend URI scheme for nodes retrieved from DNS SRV records (http or https)")
	flag.StringVar(&srvDomain, "srv-domain", "", "the name of the resource record")
	flag.StringVar(&srvRecord, "srv-record", "", "the SRV record to search for backends nodes. Example: _etcd-client._tcp.example.com")
	flag.BoolVar(&syncOnly, "sync-only", false, "sync without check_cmd and reload_cmd")
	flag.StringVar(&authType, "auth-type", "", "Vault auth backend type to use (only used with -backend=vault)")
	flag.StringVar(&appID, "app-id", "", "Vault app-id to use with the app-id backend (only used with -backend=vault and auth-type=app-id)")
	flag.StringVar(&userID, "user-id", "", "Vault user-id to use with the app-id backend (only used with -backend=value and auth-type=app-id)")
	flag.StringVar(&table, "table", "", "the name of the DynamoDB table (only used with -backend=dynamodb)")
	flag.StringVar(&username, "username", "", "the username to authenticate as (only used with vault and etcd backends)")
	flag.StringVar(&password, "password", "", "the password to authenticate with (only used with vault and etcd backends)")
	flag.BoolVar(&watch, "watch", false, "enable watch support")
	// 2nd backend config
	flag.StringVar(&authToken2, "auth-token2", "", "Auth bearer token to use for 2nd backend")
	flag.StringVar(&backend2, "backend2", "etcd", "2nd backend to use")
	flag.BoolVar(&basicAuth2, "basic-auth2", false, "Use Basic Auth to authenticate (only used with -backend=etcd) to use for 2nd backend")
	flag.StringVar(&clientCaKeys2, "client-ca-keys2", "", "client ca keys to use for 2nd backend")
	flag.StringVar(&clientCert2, "client-cert2", "", "the client cert to use for 2nd backend")
	flag.StringVar(&clientKey2, "client-key2", "", "the client key to use for 2nd backend")
	flag.Var(&nodes2, "node2", "list of 2nd backend nodes ")
	flag.StringVar(&prefix2, "prefix2", "", "key path prefix to use for 2nd backend")
	flag.StringVar(&authType2, "auth-type2", "", "Vault 2nd auth backend type to use (only used with -backend2=vault)")
	flag.StringVar(&appID2, "app-id2", "", "Vault app-id2 to use with the app-id2 on 2nd backend (only used with -backend2=vault and auth-type2=app-id)")
	flag.StringVar(&userID2, "user-id2", "", "Vault user-id to use with the 2nd app-id backend (only used with -backend2=value and auth-type2=app-id)")
	flag.StringVar(&table2, "table2", "", "the name of the DynamoDB table (only used with -backend2=dynamodb) to use for 2nd backend")
	flag.StringVar(&username2, "username2", "", "the username to authenticate as (only used with vault and etcd 2nd backends ) to use for 2nd backend")
	flag.StringVar(&password2, "password2", "", "the password to authenticate with (only used with vault and etcd 2nd backends) to use for 2nd backend")
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

	if config.SRVDomain != "" && config.SRVRecord == "" {
		config.SRVRecord = fmt.Sprintf("_%s._tcp.%s.", config.Backend, config.SRVDomain)
	}

	// Update BackendNodes from SRV records.
	if config.Backend != "env" && config.SRVRecord != "" {
		log.Info("SRV record set to " + config.SRVRecord)
		srvNodes, err := getBackendNodesFromSRV(config.SRVRecord, config.Scheme)
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

	// 2nd backend defaults if available
	if len(config.Backend2) > 0 && len(config.BackendNodes2) == 0 {
		switch config.Backend2 {
		case "consul":
			config.BackendNodes2 = []string{"127.0.0.1:8500"}
		case "etcd":
			peerstr := os.Getenv("ETCDCTL_PEERS")
			if len(peerstr) > 0 {
				config.BackendNodes2 = strings.Split(peerstr, ",")
			} else {
				config.BackendNodes2 = []string{"http://127.0.0.1:4001"}
			}
		case "redis":
			config.BackendNodes2 = []string{"127.0.0.1:6379"}
		case "zookeeper":
			config.BackendNodes2 = []string{"127.0.0.1:2181"}
		}

		// Initialize the storage client
		log.Info("2nd Backend set to " + config.Backend2)
	}


	if config.Watch {
		unsupportedBackends := map[string]bool{
			"redis":    true,
			"dynamodb": true,
			"rancher":  true,
		}

		if unsupportedBackends[config.Backend] {
			log.Info(fmt.Sprintf("Watch is not supported for backend %s. Exiting...", config.Backend))
			os.Exit(1)
		}

		// warning for 2nd backend
		if len(config.Backend2) > 0 && unsupportedBackends[config.Backend2] {
			log.Info(fmt.Sprintf("Watch is not supported for 2nd backend %s. Exiting...", config.Backend2))
			os.Exit(1)
		}
	}

	if config.Backend == "dynamodb" && config.Table == "" {
		return errors.New("No DynamoDB table configured for 1st backend")
	}

	if config.Backend2 == "dynamodb" && config.Table2 == "" {
		return errors.New("No DynamoDB table configured for 2nd backend")
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
		Username:     config.Username,
		AppID:        config.AppID,
		UserID:       config.UserID,
	}

	//2nd backend configuration
	if len(config.Backend2) > 0 {
		backendsConfig2 = backends.Config{
			AuthToken:    config.AuthToken2,
			AuthType:     config.AuthType2,
			Backend:      config.Backend2,
			BasicAuth:    config.BasicAuth2,
			ClientCaKeys: config.ClientCaKeys2,
			ClientCert:   config.ClientCert2,
			ClientKey:    config.ClientKey2,
			BackendNodes: config.BackendNodes2,
			Password:     config.Password2,
			Scheme:       config.Scheme,
			Table:        config.Table2,
			Username:     config.Username2,
			AppID:        config.AppID2,
			UserID:       config.UserID2,
		}
	}

	// Template configuration.
	templateConfig = template.Config{
		ConfDir:       config.ConfDir,
		ConfigDir:     filepath.Join(config.ConfDir, "conf.d"),
		KeepStageFile: keepStageFile,
		Noop:          config.Noop,
		Prefix:        config.Prefix,
		Prefix2:       config.Prefix2,
		SyncOnly:      config.SyncOnly,
		TemplateDir:   filepath.Join(config.ConfDir, "templates"),
	}
	return nil
}

func getBackendNodesFromSRV(record, scheme string) ([]string, error) {
	nodes := make([]string, 0)

	// Ignore the CNAME as we don't need it.
	_, addrs, err := net.LookupSRV("", "", record)
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
	// 2nd backend
	cakeys2 := os.Getenv("CONFD_CLIENT_CAKEYS2")
	if len(cakeys2) > 0 {
		config.ClientCaKeys2 = cakeys2
	}

	cert2 := os.Getenv("CONFD_CLIENT_CERT2")
	if len(cert2) > 0 {
		config.ClientCert2 = cert2
	}

	key2 := os.Getenv("CONFD_CLIENT_KEY2")
	if len(key2) > 0 {
		config.ClientKey2 = key2
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
	case "srv-domain":
		config.SRVDomain = srvDomain
	case "srv-record":
		config.SRVRecord = srvRecord
	case "sync-only":
		config.SyncOnly = syncOnly
	case "table":
		config.Table = table
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
	// 2nd backend config
	case "auth-token2":
		config.AuthToken2 = authToken2
	case "auth-type2":
		config.AuthType2 = authType2
	case "backend2":
		config.Backend2 = backend2
	case "basic-auth2":
		config.BasicAuth2 = basicAuth2
	case "client-cert2":
		config.ClientCert2 = clientCert2
	case "client-key2":
		config.ClientKey2 = clientKey2
	case "client-ca-keys2":
		config.ClientCaKeys2 = clientCaKeys2
	case "node2":
		config.BackendNodes2 = nodes2
	case "password2":
		config.Password2 = password2
	case "prefix2":
		config.Prefix2 = prefix2
	case "table2":
		config.Table2 = table2
	case "username2":
		config.Username2 = username2
	case "app-id2":
		config.AppID2 = appID2
	case "user-id2":
		config.UserID2 = userID2
	}
}
