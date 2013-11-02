// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

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
)

func init() {
	flag.Var(&nodes, "n", "list of etcd nodes")
	flag.StringVar(&confdir, "c", "/etc/confd", "confd config directory")
	flag.IntVar(&interval, "i", 600, "etcd polling interval")
	flag.StringVar(&prefix, "p", "/", "etcd key path prefix")
	flag.StringVar(&clientCert, "cert", "", "the client cert")
	flag.StringVar(&clientKey, "key", "", "the client key")
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
	if config.Confd.SRVDomain != "" {
		hostUris := make([]string, 0)
		scheme := "http"
		if config.Confd.ConnectSecure {
			scheme = "https"
		}
		etcdHosts, err := GetEtcdHostsFromSRV(config.Confd.SRVDomain)
		if err != nil {
			return errors.New("Cannot get etcd hosts from SRV records " + err.Error())
		}
		for _, h := range etcdHosts {
			uri := fmt.Sprintf("%s://%s:%d", scheme, h.Hostname, h.Port)
			hostUris = append(hostUris, uri)
		}
		config.Confd.EtcdNodes = hostUris
	}
	return nil
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
			ConfDir:   "/etc/confd",
			Interval:  600,
			Prefix:    "/",
			EtcdNodes: []string{"http://127.0.0.1:4001"},
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
	}
}

// overrideConfig iterates through each flag set on the command line and
// overrides corresponding configuration settings.
func overrideConfig() {
	flag.Visit(override)
}
