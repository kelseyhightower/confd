package etcd

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/confd/confd/config"
	"path/filepath"
	"strings"
)

var machines = []string{
	"http://127.0.0.1:4001",
}
var prefix string = "/"

func GetValues(keys []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	c := etcd.NewClient()
	success := c.SetCluster(config.EtcdNodes())
	if !success {
		return vars, nil
	}
	r := strings.NewReplacer("/", "_")
	for _, key := range keys {
		values, err := c.Get(filepath.Join(config.Prefix(), key))
		if err != nil {
			return vars, err
		}
		for _, v := range values {
			key := strings.TrimPrefix(v.Key, config.Prefix())
			new_key := r.Replace(key)
			vars[new_key] = v.Value
		}
	}
	return vars, nil
}
