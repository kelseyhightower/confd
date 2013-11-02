package main

import (
	"net"
	"os"
	"strings"
)

// etcdHost
type etcdHost struct {
	Hostname string
	Port     uint16
}

// IsFileExist reports whether path exits.
func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
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
