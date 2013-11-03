// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/log"
	"strconv"
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

// Nodes is a custom flag Var representing a list of etcd nodes. We use a custom
// Var to allow us to define more than one etcd node from the command line, and
// collect the results in a single value.
type Nodes []string

func (n *Nodes) String() string {
	return fmt.Sprintf("%d", *n)
}

// Set appends the node to the etcd node list.
func (n *Nodes) Set(node string) error {
	*n = append(*n, node)
	return nil
}

// Config represents the confd configuration settings.
type Config struct {
	Confd confd
}

// confd represents the parsed configuration settings.
type confd struct {
	ConfDir       string
	ClientCert    string `toml:"client_cert"`
	ClientKey     string `toml:"client_key"`
	ConnectSecure bool   `toml:"connect_secure"`
	Interval      int
	Prefix        string
	EtcdNodes     []string `toml:"etcd_nodes"`
	EtcdScheme    string   `toml:"etcd_scheme"`
	Noop          bool     `toml:"noop"`
	SRVDomain     string   `toml:"srv_domain"`
}

// loadConfig initializes the confd configuration by first setting defaults,
// then overriding setting from the confd config file, and finally overriding
// settings from flags set on the command line.
// It returns an error if any.
func loadConfig(path string) error {
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

// Prefix returns the etcd key prefix to use when querying etcd.
func Prefix() string {
	return config.Confd.Prefix
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
