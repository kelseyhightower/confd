// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package config

import (
	"testing"
	"github.com/kelseyhightower/confd/log"
)

func TestLoadConfig(t *testing.T) {
	log.SetQuiet(true)
	var expected = struct {
		clientCert  string
		clientKey   string
		configDir   string
		etcdNodes   []string
		interval    int
		prefix      string
		templateDir string
	}{
		"", "", "/etc/confd/conf.d", []string{"http://127.0.0.1:4001"},
		600, "/", "/etc/confd/templates",
	}
	LoadConfig("")
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
	p := Prefix()
	if p != expected.prefix {
		t.Errorf("Expected default prefix = %s, got %s", expected.prefix, p)
	}
	td := TemplateDir()
	if td != expected.templateDir {
		t.Errorf("Expected default templateDir = %s, got %s", expected.templateDir, td)
	}
}
