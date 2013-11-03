// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/log"
)

var (
	config     Config
	nodes      Nodes
	confdir    string
	interval   int
	prefix     string
	clientCert string
	clientKey  string
	srvDomain  string
	noop       bool
	etcdScheme string
)

// etcdHost
type etcdHost struct {
	Hostname string
	Port     uint16
}

func init() {
	flag.Var(&nodes, "n", "list of etcd nodes")
	flag.StringVar(&confdir, "c", "/etc/confd", "confd config directory")
	flag.IntVar(&interval, "i", 600, "etcd polling interval")
	flag.StringVar(&prefix, "p", "/", "etcd key path prefix")
	flag.StringVar(&clientCert, "cert", "", "the client cert")
	flag.StringVar(&clientKey, "key", "", "the client key")
	flag.StringVar(&etcdScheme, "etcd-scheme", "http", "the etcd URI scheme. (http or https)")
	flag.StringVar(&srvDomain, "srv-domain", "", "the domain to query for the etcd SRV record, i.e. example.com")
	flag.BoolVar(&noop, "noop", false, "only show pending changes, don't sync configs.")
}


// Config represents the confd configuration settings.
type Config struct {
	Confd confd
}

// confd represents the parsed configuration settings.
type confd struct {
	ConfDir    string
	ClientCert string `toml:"client_cert"`
	ClientKey  string `toml:"client_key"`
	Interval   int
	Prefix     string
	EtcdNodes  []string `toml:"etcd_nodes"`
	EtcdScheme string   `toml:"etcd_scheme"`
	Noop       bool     `toml:"noop"`
	SRVDomain  string   `toml:"srv_domain"`
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
		if err := loadConfFile(path); err != nil {
			return err
		}
	}
	overrideConfig()
	if !isValidateEtcdScheme(config.Confd.EtcdScheme) {
		return errors.New("Invalid etcd scheme: " + config.Confd.EtcdScheme)
	}
	err := setEtcdHosts()
	if err != nil {
		return err
	}
	return nil
}

func setEtcdHosts() error {
	scheme := config.Confd.EtcdScheme
	hosts := make([]string, 0)
	// If a domain name is given then lookup the etcd SRV record, and override
	// all other etcd node settings.
	if config.Confd.SRVDomain != "" {
		etcdHosts, err := GetEtcdHostsFromSRV(config.Confd.SRVDomain)
		if err != nil {
			return errors.New("Cannot get etcd hosts from SRV records " + err.Error())
		}
		for _, h := range etcdHosts {
			uri := formatEtcdHostURI(scheme, h.Hostname, strconv.FormatUint(uint64(h.Port), 10))
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
			hosts = append(hosts, formatEtcdHostURI(etcdURL.Scheme, host, port))
			continue
		}
		// At this point node is not an etcd URL, i.e. http://etcd.example.com:4001,
		// but a host:port string, i.e. etcd.example.com:4001
		host, port, err := net.SplitHostPort(node)
		if err != nil {
			return err
		}
		hosts = append(hosts, formatEtcdHostURI(scheme, host, port))
	}
	config.Confd.EtcdNodes = hosts
	return nil
}

func formatEtcdHostURI(scheme, host, port string) string {
	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}

func isValidateEtcdScheme(scheme string) bool {
	if scheme == "http" || scheme == "https" {
		return true
	}
	return false
}

// ConfigDir returns the path to the confd config dir.
func ConfigDir() string {
	return filepath.Join(config.Confd.ConfDir, "conf.d")
}

// SetConfDir.
func SetConfDir(path string) {
	config.Confd.ConfDir = path
}

// ClientCert returns the path to the client cert.
func ClientCert() string {
	return config.Confd.ClientCert
}

// ClientKey returns the path to the client key.
func ClientKey() string {
	return config.Confd.ClientKey
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

// Noop
func Noop() bool {
	return config.Confd.Noop
}

// SetNoop.
func SetNoop(enabled bool) {
	config.Confd.Noop = enabled
}

// Prefix returns the etcd key prefix to use when querying etcd.
func Prefix() string {
	return config.Confd.Prefix
}

// SetPrefix
func SetPrefix(prefix string) {
	config.Confd.Prefix = prefix
}

// TemplateDir returns the path to the directory of config file templates.
func TemplateDir() string {
	return filepath.Join(config.Confd.ConfDir, "templates")
}

// SRVDomain return the domain name used to query etcd SRV records.
func SRVDomain() string {
	return config.Confd.SRVDomain
}

func setDefaults() {
	config = Config{
		Confd: confd{
			ConfDir:    "/etc/confd",
			Interval:   600,
			Prefix:     "/",
			EtcdNodes:  []string{"127.0.0.1:4001"},
			EtcdScheme: "http",
		},
	}
}

// loadConfFile sets the etcd configuration settings from a file.
func loadConfFile(path string) error {
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return err
	}
	return nil
}

// override sets configuration settings based on values passed in through
// command line flags; overwriting current values.
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
	case "cert":
		config.Confd.ClientCert = clientCert
	case "key":
		config.Confd.ClientKey = clientKey
	case "noop":
		config.Confd.Noop = noop
	case "srv-domain":
		config.Confd.SRVDomain = srvDomain
	case "etcd-scheme":
		config.Confd.EtcdScheme = etcdScheme
	}
}

// overrideConfig iterates through each flag set on the command line and
// overrides corresponding configuration settings.
func overrideConfig() {
	flag.Visit(override)
}

// GetEtcdHostsFromSRV returns a list of etcHost.
func GetEtcdHostsFromSRV(domain string) ([]*etcdHost, error) {
	addrs, err := lookupEtcdSRV(domain)
	if err != nil {
		return nil, err
	}
	etcdHosts := etcdHostsFromSRV(addrs)
	return etcdHosts, nil
}

// lookupEtcdSrv tries to resolve an SRV query for the etcd service for the
// specified domain.
//
// lookupEtcdSRV constructs the DNS name to look up following RFC 2782.
// That is, it looks up _etcd._tcp.domain.
func lookupEtcdSRV(domain string) ([]*net.SRV, error) {
	// Ignore the CNAME as we don't need it.
	_, addrs, err := net.LookupSRV("etcd", "tcp", domain)
	if err != nil {
		return addrs, err
	}
	return addrs, nil
}

// etcdHostsFromSRV converts an etcd SRV record to a list of etcdHost.
func etcdHostsFromSRV(addrs []*net.SRV) []*etcdHost {
	hosts := make([]*etcdHost, 0)
	for _, srv := range addrs {
		hostname := strings.TrimRight(srv.Target, ".")
		hosts = append(hosts, &etcdHost{Hostname: hostname, Port: srv.Port})
	}
	return hosts
}
