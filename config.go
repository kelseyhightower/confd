package main

import (
	"errors"
	"flag"
	"fmt"
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

type TemplateConfig = template.Config
type BackendsConfig = backends.Config

// A Config structure is used to configure confd.
type Config struct {
	TemplateConfig
	BackendsConfig
	Interval     int    `toml:"interval"`
	SRVDomain    string `toml:"srv_domain"`
	SRVRecord    string `toml:"srv_record"`
	LogLevel     string `toml:"log-level"`
	Watch        bool   `toml:"watch"`
	PrintVersion bool
	ConfigFile   string
	OneTime      bool
	SyncFirst    bool
}

var config Config

func init() {
	flag.StringVar(&config.AuthToken, "auth-token", "", "Auth bearer token to use")
	flag.StringVar(&config.Backend, "backend", "etcd", "backend to use")
	flag.BoolVar(&config.BasicAuth, "basic-auth", false, "Use Basic Auth to authenticate (only used with -backend=consul and -backend=etcd)")
	flag.StringVar(&config.ClientCaKeys, "client-ca-keys", "", "client ca keys")
	flag.StringVar(&config.ClientCert, "client-cert", "", "the client cert")
	flag.StringVar(&config.ClientKey, "client-key", "", "the client key")
	flag.BoolVar(&config.ClientInsecure, "client-insecure", false, "Allow connections to SSL sites without certs (only used with -backend=etcd)")
	flag.StringVar(&config.ConfDir, "confdir", "/etc/confd", "confd conf directory")
	flag.StringVar(&config.ConfigFile, "config-file", "/etc/confd/confd.toml", "the confd config file")
	flag.Var(&config.YAMLFile, "file", "the YAML file to watch for changes (only used with -backend=file)")
	flag.StringVar(&config.Filter, "filter", "*", "files filter (only used with -backend=file)")
	flag.IntVar(&config.Interval, "interval", 600, "backend polling interval")
	flag.BoolVar(&config.KeepStageFile, "keep-stage-file", false, "keep staged files")
	flag.StringVar(&config.LogLevel, "log-level", "", "level which confd should log messages")
	flag.Var(&config.BackendNodes, "node", "list of backend nodes")
	flag.BoolVar(&config.Noop, "noop", false, "only show pending changes")
	flag.BoolVar(&config.OneTime, "onetime", false, "run once and exit")
	flag.StringVar(&config.Prefix, "prefix", "", "key path prefix")
	flag.BoolVar(&config.PrintVersion, "version", false, "print version and exit")
	flag.StringVar(&config.Scheme, "scheme", "http", "the backend URI scheme for nodes retrieved from DNS SRV records (http or https)")
	flag.StringVar(&config.SRVDomain, "srv-domain", "", "the name of the resource record")
	flag.StringVar(&config.SRVRecord, "srv-record", "", "the SRV record to search for backends nodes. Example: _etcd-client._tcp.example.com")
	flag.BoolVar(&config.SyncOnly, "sync-only", false, "sync without check_cmd and reload_cmd")
	flag.StringVar(&config.AuthType, "auth-type", "", "Vault auth backend type to use (only used with -backend=vault)")
	flag.StringVar(&config.AppID, "app-id", "", "Vault app-id to use with the app-id backend (only used with -backend=vault and auth-type=app-id)")
	flag.StringVar(&config.UserID, "user-id", "", "Vault user-id to use with the app-id backend (only used with -backend=value and auth-type=app-id)")
	flag.StringVar(&config.RoleID, "role-id", "", "Vault role-id to use with the AppRole, Kubernetes backends (only used with -backend=vault and either auth-type=app-role or auth-type=kubernetes)")
	flag.StringVar(&config.SecretID, "secret-id", "", "Vault secret-id to use with the AppRole backend (only used with -backend=vault and auth-type=app-role)")
	flag.StringVar(&config.Path, "path", "", "Vault mount path of the auth method (only used with -backend=vault)")
	flag.StringVar(&config.Table, "table", "", "the name of the DynamoDB table (only used with -backend=dynamodb)")
	flag.StringVar(&config.Separator, "separator", "", "the separator to replace '/' with when looking up keys in the backend, prefixed '/' will also be removed (only used with -backend=redis)")
	flag.StringVar(&config.Username, "username", "", "the username to authenticate as (only used with vault and etcd backends)")
	flag.StringVar(&config.Password, "password", "", "the password to authenticate with (only used with vault and etcd backends)")
	flag.BoolVar(&config.Watch, "watch", false, "enable watch support")
	flag.BoolVar(&config.SyncFirst, "sync-first", false, "sync template first")
}

// initConfig initializes the confd configuration by first setting defaults,
// then overriding settings from the confd config file, then overriding
// settings from environment variables, and finally overriding
// settings from flags set on the command line.
// It returns an error if any.
func initConfig() error {
	_, err := os.Stat(config.ConfigFile)
	if os.IsNotExist(err) {
		log.Debug("Skipping confd config file.")
	} else {
		log.Debug("Loading " + config.ConfigFile)
		configBytes, err := os.ReadFile(config.ConfigFile)
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
		return errors.New("no DynamoDB table configured")
	}
	config.ConfigDir = filepath.Join(config.ConfDir, "conf.d")
	config.TemplateDir = filepath.Join(config.ConfDir, "templates")
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

func processEnv() {
	cakeys := os.Getenv("CONFD_CLIENT_CAKEYS")
	if len(cakeys) > 0 && config.ClientCaKeys == "" {
		config.ClientCaKeys = cakeys
	}

	cert := os.Getenv("CONFD_CLIENT_CERT")
	if len(cert) > 0 && config.ClientCert == "" {
		config.ClientCert = cert
	}

	key := os.Getenv("CONFD_CLIENT_KEY")
	if len(key) > 0 && config.ClientKey == "" {
		config.ClientKey = key
	}
}
