package confd

import (
	"reflect"
	"testing"

	"github.com/kelseyhightower/confd/log"
)

func TestInitConfigDefaultConfig(t *testing.T) {
	log.SetLevel("warn")
	want := Config{
		Backend:      "etcd",
		BackendNodes: []string{"http://127.0.0.1:4001"},
		ClientCaKeys: "",
		ClientCert:   "",
		ClientKey:    "",
		ConfDir:      "/etc/confd",
		Interval:     600,
		Noop:         false,
		Prefix:       "/",
		SRVDomain:    "",
		Scheme:       "http",
	}
	if err := InitConfig(); err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(want, Cfg) {
		t.Errorf("initConfig() = %v, want %v", Cfg, want)
	}
}
