package etcdutil

import (
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"path/filepath"
	"strings"
)

var replacer = strings.NewReplacer("/", "_")

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewEtcdClient(machines []string, cert, key string) (*etcd.Client, error) {
	c := etcd.NewClient(machines)
	if cert != "" && key != "" {
		_, err := c.SetCertAndKey(cert, key)
		if err != nil {
			return c, err
		}
	}
	success := c.SetCluster(machines)
	if !success {
		return c, errors.New("cannot connect to etcd cluster: " + strings.Join(machines, ","))
	}
	return c, nil
}

type EtcdClient interface {
	Get(key string) ([]*etcd.Response, error)
}

// GetValues queries etcd for keys prefixed by prefix.
// Etcd paths (keys) are translated into names more suitable for use in
// templates. For example if prefix where set to '/production' and one of the
// keys where '/nginx/port'; the prefixed '/production/nginx/port' key would
// be quired for. If the value for the prefixed key where 80, the returned map
// would contain the entry vars["nginx_port"] = "80".
func GetValues(c EtcdClient, prefix string, keys []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	for _, key := range keys {
		err := etcdWalk(c, filepath.Join(prefix, key), prefix, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

// etcdWalk recursively descends etcd paths, updating vars.
func etcdWalk(c EtcdClient, key string, prefix string, vars map[string]interface{}) error {
	values, err := c.Get(key)
	if err != nil {
		return err
	}
	for _, v := range values {
		if !v.Dir {
			key := pathToKey(v.Key, prefix)
			vars[key] = v.Value
		} else {
			etcdWalk(c, v.Key, prefix, vars)
		}
	}
	return nil
}

// pathToKey translates etcd key paths into something more suitable for use
// in Golang templates. Turn /prefix/key/subkey into key_subkey.
func pathToKey(key, prefix string) string {
	key = strings.TrimPrefix(key, prefix)
	key = strings.TrimPrefix(key, "/")
	return replacer.Replace(key)
}
