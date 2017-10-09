package main

import (
	"reflect"
	"testing"

	"github.com/kelseyhightower/confd/log"
)

func TestInitConfigDefaultConfig(t *testing.T) {
	log.SetLevel("warn")
	want := Config{
		Backend:       "etcd",
		BackendNodes:  []string{"http://127.0.0.1:4001"},
		ClientCaKeys:  "",
		ClientCert:    "",
		ClientKey:     "",
		ConfDir:       "/etc/confd",
		Interval:      600,
		Noop:          false,
		Prefix:        "",
		SRVDomain:     "",
		Scheme:        "http",
		SecretKeyring: "",
		Table:         "",
	}
	if err := initConfig(); err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(want, config) {
		t.Errorf("initConfig() = %v, want %v", config, want)
	}
}
