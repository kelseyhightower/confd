// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"reflect"
	"testing"

	"github.com/kelseyhightower/confd/log"
)

func TestInitConfigDefaultConfig(t *testing.T) {
	log.SetQuiet(true)
	want := Config{
		Backend:      "etcd",
		BackendNodes: []string{"127.0.0.1:4001"},
		ClientCaKeys: "",
		ClientCert:   "",
		ClientKey:    "",
		ConfDir:      "/etc/confd",
		Debug:        false,
		Interval:     600,
		Noop:         false,
		Prefix:       "",
		Quiet:        false,
		SRVDomain:    "",
		Scheme:       "http",
		Verbose:      false,
	}
	if err := initConfig(); err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(want, config) {
		t.Errorf("initConfig() = %v, want %v", config, want)
	}
}
