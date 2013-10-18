package main

import (
	"testing"
)

func TestInitConfig(t *testing.T) {
	var expected = struct {
		clientCert  string
		clientKey   string
		configDir   string
		etcdNodes   []string
		interval    int
		onetime     bool
		prefix      string
		templateDir string
	}{
		"", "", "/etc/confd/conf.d", []string{"http://127.0.0.1:4001"},
		600, false, "/", "/etc/confd/templates",
	}
	InitConfig()
	cc := ClientCert()
	if cc != expected.clientCert {
		t.Errorf("Expected default clientCert = %s, got %s", expected.clientCert, cc)
	}
	ck := ClientKey()
	if ck != expected.clientKey {
		t.Errorf("Expected default clientKey = %s, got %s", expected.clientKey, ck)
	}
	cd := ConfigDir()
	if cd != expected.configDir {
		t.Errorf("Expected default configDir = %s, got %s", expected.configDir, cd)
	}
	en := EtcdNodes()
	if en[0] != expected.etcdNodes[0] {
		t.Errorf("Expected default etcdNodes = %v, got %v", expected.etcdNodes, en)
	}
	i := Interval()
	if i != expected.interval {
		t.Errorf("Expected default interval = %d, got %d", expected.interval, i)
	}
	ot := Onetime()
	if ot != expected.onetime {
		t.Errorf("Expected default onetime = %v, got %v", expected.onetime, ot)
	}
	p := Prefix()
	if p != expected.prefix {
		t.Errorf("Expected default prefix = %s, got %s", expected.prefix, p)
	}
	td := TemplateDir()
	if td != expected.templateDir {
		t.Errorf("Expected default templateDir = %s, got %s", expected.templateDir, td)
	}
}
