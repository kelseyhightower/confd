package main

import (
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"path/filepath"
	"strings"
)

// newEtcdClient returns an *etcd.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func newEtcdClient(machines []string, cert, key string) (*etcd.Client, error) {
	c := etcd.NewClient()
	if cert != "" {
		_, err := c.SetCertAndKey(cert, key)
		if err != nil {
			return c, err
		}
	}
	success := c.SetCluster(machines)
	if !success {
		return c, errors.New("cannot connect to etcd cluster")
	}
	return c, nil
}

// getValues queries a etcd for keys prefixed by prefix.
// Etcd paths (keys) are translated into names more suitable for use in
// templates. For example if prefix where set to '/production' and one of the
// keys where '/nginx/port'; the prefixed '/production/nginx/port' key would
// be quired for. If the value for the prefixed key where 80, the returned map
// would contain the entry vars["nginx_port"] = "80".
func getValues(c *etcd.Client, prefix string, keys []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	r := strings.NewReplacer("/", "_")
	for _, key := range keys {
		values, err := c.Get(filepath.Join(prefix, key))
		if err != nil {
			return vars, err
		}
		for _, v := range values {
			// Translate the prefixed etcd path into something more suitable
			// for use in a template. Turn /prefix/key/subkey into key_subkey.
			key := strings.TrimPrefix(v.Key, prefix)
			key = strings.TrimPrefix(key, "/")
			new_key := r.Replace(key)
			vars[new_key] = v.Value
		}
	}
	return vars, nil
}
