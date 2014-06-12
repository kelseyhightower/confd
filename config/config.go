// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package config

import (
	"errors"
	"flag"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/log"
)

var (
	backend      string
	clientCert   string
	clientKey    string
	clientCaKeys string
	config       Config // holds the global confd config.
	confdir      string
	consul       bool
	consulAddr   string
	debug        bool
	etcdNodes    Nodes
	etcdScheme   string
	interval     int
	noop         bool
	prefix       string
	quiet        bool
	srvDomain    string
	verbose      bool
)

// Config represents the confd configuration settings.
type Config struct {
	Confd confd
}

// confd represents the parsed configuration settings.
type confd struct {
	Backend      string   `toml:"backend"`
	Debug        bool     `toml:"debug"`
	ClientCert   string   `toml:"client_cert"`
	ClientKey    string   `toml:"client_key"`
	ClientCaKeys string   `toml:"client_cakeys"`
	ConfDir      string   `toml:"confdir"`
	Consul       bool     `toml:"consul"`
	ConsulAddr   string   `toml:"consul_addr"`
	EtcdNodes    []string `toml:"etcd_nodes"`
	EtcdScheme   string   `toml:"etcd_scheme"`
	Interval     int      `toml:"interval"`
	Noop         bool     `toml:"noop"`
	Prefix       string   `toml:"prefix"`
	Quiet        bool     `toml:"quiet"`
	SRVDomain    string   `toml:"srv_domain"`
	Verbose      bool     `toml:"verbose"`
}

func init() {
	flag.StringVar(&backend, "backend", "", "backend to use")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.StringVar(&clientCert, "client-cert", "", "the client cert")
	flag.StringVar(&clientKey, "client-key", "", "the client key")
	flag.StringVar(&clientCaKeys, "client-ca-keys", "", "client ca keys")
	flag.StringVar(&confdir, "confdir", "/etc/confd", "confd conf directory")
	flag.BoolVar(&consul, "consul", false, "specified to enable use of Consul")
	flag.StringVar(&consulAddr, "consul-addr", "", "address of Consul HTTP interface")
	flag.Var(&etcdNodes, "node", "list of etcd nodes")
	flag.StringVar(&etcdScheme, "etcd-scheme", "http", "the etcd URI scheme. (http or https)")
	flag.IntVar(&interval, "interval", 600, "etcd polling interval")
	flag.BoolVar(&noop, "noop", false, "only show pending changes, don't sync configs.")
	flag.StringVar(&prefix, "prefix", "/", "etcd key path prefix")
	flag.BoolVar(&quiet, "quiet", false, "enable quiet logging. Only error messages are printed.")
	flag.StringVar(&srvDomain, "srv-domain", "", "the domain to query for the etcd SRV record, i.e. example.com")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose logging")
}

// LoadConfig initializes the confd configuration by first setting defaults,
// then overriding setting from the confd config file, and finally overriding
// settings from flags set on the command line.
// It returns an error if any.
func LoadConfig(path string) error {
	setDefaults()
	if path == "" {
		log.Warning("Skipping confd config file.")
	} else {
		log.Debug("Loading " + path)

		configBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}
	processFlags()
	if !isValidateEtcdScheme(config.Confd.EtcdScheme) {
		return errors.New("Invalid etcd scheme: " + config.Confd.EtcdScheme)
	}
	err := setEtcdHosts()
	if err != nil {
		return err
	}
	return nil
}

func Backend() string {
	return config.Confd.Backend
}

// Debug reports whether debug mode is enabled.
func Debug() bool {
	return config.Confd.Debug
}

// ClientCert returns the client cert path.
func ClientCert() string {
	return config.Confd.ClientCert
}

// ClientKey returns the client key path.
func ClientKey() string {
	return config.Confd.ClientKey
}

// ClientCaKeys returns the client CA certificates
func ClientCaKeys() string {
	return config.Confd.ClientCaKeys
}

// ConfDir returns the path to the confd config dir.
func ConfDir() string {
	return config.Confd.ConfDir
}

// ConfigDir returns the path to the confd config dir.
func ConfigDir() string {
	return filepath.Join(config.Confd.ConfDir, "conf.d")
}

// Consul returns if we should use Consul
func Consul() bool {
	return config.Confd.Consul
}

// ConsulAddr returns the address of the consul node
func ConsulAddr() string {
	return config.Confd.ConsulAddr
}

// EtcdNodes returns a list of etcd node url strings.
// For example: ["http://203.0.113.30:4001"]
func EtcdNodes() []string {
	return config.Confd.EtcdNodes
}

// Interval returns the number of seconds to wait between configuration runs.
func Interval() int {
	return config.Confd.Interval
}

// Noop reports whether noop mode is enabled.
func Noop() bool {
	return config.Confd.Noop
}

// Prefix returns the etcd key prefix to use when querying etcd.
func Prefix() string {
	return config.Confd.Prefix
}

// Quiet reports whether quiet mode is enabled.
func Quiet() bool {
	return config.Confd.Quiet
}

// Verbose reports whether verbose mode is enabled.
func Verbose() bool {
	return config.Confd.Verbose
}

// SetConfDir sets the confd conf dir.
func SetConfDir(path string) {
	config.Confd.ConfDir = path
}

// SetNoop sets noop.
func SetNoop(enabled bool) {
	config.Confd.Noop = enabled
}

// SetPrefix sets the key prefix.
func SetPrefix(prefix string) {
	config.Confd.Prefix = prefix
}

// SRVDomain returns the domain name used in etcd SRV record lookups.
func SRVDomain() string {
	return config.Confd.SRVDomain
}

// TemplateDir returns the template directory path.
func TemplateDir() string {
	return filepath.Join(config.Confd.ConfDir, "templates")
}

func setDefaults() {
	config = Config{
		Confd: confd{
			ConfDir:    "/etc/confd",
			ConsulAddr: "127.0.0.1:8500",
			Interval:   600,
			Prefix:     "/",
			EtcdNodes:  []string{"127.0.0.1:4001"},
			EtcdScheme: "http",
		},
	}
}

// setEtcdHosts.
func setEtcdHosts() error {
	scheme := config.Confd.EtcdScheme
	hosts := make([]string, 0)
	// If a domain name is given then lookup the etcd SRV record, and override
	// all other etcd node settings.
	if config.Confd.SRVDomain != "" {
		log.Info("SRV domain set to " + config.Confd.SRVDomain)
		etcdHosts, err := getEtcdHostsFromSRV(config.Confd.SRVDomain)
		if err != nil {
			return errors.New("Cannot get etcd hosts from SRV records " + err.Error())
		}
		for _, h := range etcdHosts {
			uri := formatEtcdHostURL(scheme, h.Hostname, strconv.FormatUint(uint64(h.Port), 10))
			hosts = append(hosts, uri)
		}
		config.Confd.EtcdNodes = hosts
		return nil
	}
	// No domain name was given, so just process the etcd node list.
	// An etcdNode can be a URL, http://etcd.example.com:4001, or a host, etcd.example.com:4001.
	for _, node := range config.Confd.EtcdNodes {
		etcdURL, err := url.Parse(node)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		if etcdURL.Scheme != "" && etcdURL.Host != "" {
			if !isValidateEtcdScheme(etcdURL.Scheme) {
				return errors.New("The etcd node list contains an invalid URL: " + node)
			}
			host, port, err := net.SplitHostPort(etcdURL.Host)
			if err != nil {
				return err
			}
			hosts = append(hosts, formatEtcdHostURL(etcdURL.Scheme, host, port))
			continue
		}
		// At this point node is not an etcd URL, i.e. http://etcd.example.com:4001,
		// but a host:port string, i.e. etcd.example.com:4001
		host, port, err := net.SplitHostPort(node)
		if err != nil {
			return err
		}
		hosts = append(hosts, formatEtcdHostURL(scheme, host, port))
	}
	config.Confd.EtcdNodes = hosts
	return nil
}

// processFlags iterates through each flag set on the command line and
// overrides corresponding configuration settings.
func processFlags() {
	flag.Visit(setConfigFromFlag)
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "backend":
		config.Confd.Backend = backend
	case "debug":
		config.Confd.Debug = debug
	case "client-cert":
		config.Confd.ClientCert = clientCert
	case "client-key":
		config.Confd.ClientKey = clientKey
	case "client-cakeys":
		config.Confd.ClientCaKeys = clientCaKeys
	case "confdir":
		config.Confd.ConfDir = confdir
	case "consul":
		config.Confd.Consul = consul
	case "consul-addr":
		config.Confd.ConsulAddr = consulAddr
	case "node":
		config.Confd.EtcdNodes = etcdNodes
	case "etcd-scheme":
		config.Confd.EtcdScheme = etcdScheme
	case "interval":
		config.Confd.Interval = interval
	case "noop":
		config.Confd.Noop = noop
	case "prefix":
		config.Confd.Prefix = prefix
	case "quiet":
		config.Confd.Quiet = quiet
	case "srv-domain":
		config.Confd.SRVDomain = srvDomain
	case "verbose":
		config.Confd.Verbose = verbose
	}
}
