// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package config

import (
	"fmt"
	"net"
	"strings"
)

// etcdHost
type etcdHost struct {
	Hostname string
	Port     uint16
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

// formatEtcdHostURL.
func formatEtcdHostURL(scheme, host, port string) string {
	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}

// getEtcdHostsFromSRV returns a list of etcHost.
func getEtcdHostsFromSRV(domain string) ([]*etcdHost, error) {
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

// isValidateEtcdScheme.
func isValidateEtcdScheme(scheme string) bool {
	if scheme == "http" || scheme == "https" {
		return true
	}
	return false
}
