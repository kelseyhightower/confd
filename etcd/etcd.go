// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

/*
Package etcd provides an interface to etcd clusters.
*/
package etcd

import (
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"path/filepath"
	"strings"
)

// GetValues queries the named set of nodes for keys prefixed by prefix.
// Etcd paths (keys) are translated into names more suitable for use in
// templates. For example if prefix where set to '/production' and one of the
// keys where '/nginx/port'; the prefixed '/production/nginx/port' key would
// be quired for. If the value for the prefixed key where 80, the returned map
// would contain the entry vars["nginx_port"] = "80".
func GetValues(prefix string, keys, nodes []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	c := etcd.NewClient()
	success := c.SetCluster(nodes)
	if !success {
		return vars, errors.New("cannot connect to etcd cluster")
	}
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
