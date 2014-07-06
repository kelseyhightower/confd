// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/log"
)

var (
	backend      string
	clientCaKeys string
	clientCert   string
	clientKey    string
	confdir      string
	config       Config // holds the global confd config.
	debug        bool
	interval     int
	nodes        Nodes
	noop         bool
	prefix       string
	quiet        bool
	scheme       string
	srvDomain    string
	verbose      bool
)

// A Config structure is used to configure confd.
type Config struct {
	Backend      string   `toml:"backend"`
	BackendNodes []string `toml:"nodes"`
	ClientCaKeys string   `toml:"client_cakeys"`
	ClientCert   string   `toml:"client_cert"`
	ClientKey    string   `toml:"client_key"`
	ConfDir      string   `toml:"confdir"`
	ConfigDir    string
	Debug        bool   `toml:"debug"`
	Interval     int    `toml:"interval"`
	Noop         bool   `toml:"noop"`
	Prefix       string `toml:"prefix"`
	Quiet        bool   `toml:"quiet"`
	SRVDomain    string `toml:"srv_domain"`
	Scheme       string `toml:"scheme"`
	TemplateDir  string
	Verbose      bool `toml:"verbose"`
}

func init() {
	flag.StringVar(&backend, "backend", "", "backend to use")
	flag.StringVar(&clientCaKeys, "client-ca-keys", "", "client ca keys")
	flag.StringVar(&clientCert, "client-cert", "", "the client cert")
	flag.StringVar(&clientKey, "client-key", "", "the client key")
	flag.StringVar(&confdir, "confdir", "/etc/confd", "confd conf directory")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.IntVar(&interval, "interval", 600, "etcd polling interval")
	flag.Var(&nodes, "node", "list of backend nodes")
	flag.BoolVar(&noop, "noop", false, "only show pending changes, don't sync configs.")
	flag.StringVar(&prefix, "prefix", "/", "key path prefix")
	flag.BoolVar(&quiet, "quiet", false, "enable quiet logging. Only error messages are printed.")
	flag.StringVar(&scheme, "backend-scheme", "http", "the backend URI scheme. (http or https)")
	flag.StringVar(&srvDomain, "srv-domain", "", "the domain to query for the backend SRV record, i.e. example.com")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose logging")
}

// LoadConfig initializes the confd configuration by first setting defaults,
// then overriding setting from the confd config file, and finally overriding
// settings from flags set on the command line.
// It returns an error if any.
func LoadConfig(path string) error {
	// Set defaults.
	config = Config{
		ConfDir:      "/etc/confd",
		Interval:     600,
		Prefix:       "/",
		BackendNodes: []string{"127.0.0.1:4001"},
	}
	// Update config from the TOML configuration file.
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
	// Update config from commandline flags.
	processFlags()

	// Update BackendNodes from SRV records.
	if config.Backend != "env" && config.SRVDomain != "" {
		log.Info("SRV domain set to " + config.SRVDomain)
		srvNodes, err := getBackendNodesFromSRV(config.Backend, config.SRVDomain)
		if err != nil {
			return errors.New("Cannot get nodes from SRV records " + err.Error())
		}
		config.BackendNodes = srvNodes
	}

	config.ConfigDir = filepath.Join(config.ConfDir, "conf.d")
	config.TemplateDir = filepath.Join(config.ConfDir, "templates")
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
		config.Backend = backend
	case "debug":
		config.Debug = debug
	case "client-cert":
		config.ClientCert = clientCert
	case "client-key":
		config.ClientKey = clientKey
	case "client-cakeys":
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
	case "quiet":
		config.Quiet = quiet
	case "scheme":
		config.Scheme = scheme
	case "srv-domain":
		config.SRVDomain = srvDomain
	case "verbose":
		config.Verbose = verbose
	}
}

func getBackendNodesFromSRV(backend, domain string) ([]string, error) {
	nodes := make([]string, 0)
	// Ignore the CNAME as we don't need it.
	_, addrs, err := net.LookupSRV(backend, "tcp", domain)
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
