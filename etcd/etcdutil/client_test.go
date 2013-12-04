// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package etcdutil

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/confd/etcd/etcdtest"
	"testing"
)

type PathToKeyTest struct {
	key, prefix, expected string
}

var pathToKeyTests = []PathToKeyTest{
	// Without prefix
	{"/nginx/port", "", "nginx_port"},
	{"/nginx/worker_processes", "", "nginx_worker_processes"},
	{"/foo/bar/mat/zoo", "", "foo_bar_mat_zoo"},
	// With prefix
	{"/prefix/nginx/port", "/prefix", "nginx_port"},
	// With prefix and trailing slash
	{"/prefix/nginx/port", "/prefix/", "nginx_port"},
}

func TestPathToKey(t *testing.T) {
	for _, pt := range pathToKeyTests {
		result := pathToKey(pt.key, pt.prefix)
		if result != pt.expected {
			t.Errorf("Expected pathToKey(%s, %s) to == %s, got %s",
				pt.key, pt.prefix, pt.expected, result)
		}
	}
}

func TestGetValues(t *testing.T) {
	// Use stub etcd client.
	c := etcdtest.NewClient()

	fooResp := &etcd.Response{
		Action: "GET",
		Node: &etcd.Node{
			Key: "/foo",
			Dir: true,
			Value: "",
			Nodes: etcd.Nodes{
				etcd.Node{Key: "/foo/one",Dir: false,Value: "one"},
				etcd.Node{Key: "foo/two", Dir: false, Value: "two"},
				etcd.Node{
					Key: "/foo/three",
					Dir: true,
					Value: "",
					Nodes: etcd.Nodes{
						etcd.Node{Key: "/foo/three/bar", Value: "three_bar", Dir: false},
					},
				},
			},
		},
	}
	quuxResp := &etcd.Response{
		Action: "GET",
		Node: &etcd.Node{Key:"/quux", Dir: false, Value: "foo"},
	}
	nginxResp := &etcd.Response{
		Action: "GET",
		Node: &etcd.Node{
			Key: "/nginx",
			Value: "",
			Dir: true,
			Nodes: etcd.Nodes{
				etcd.Node{Key: "/nginx/port", Dir: false, Value: "443"},
				etcd.Node{Key: "/nginx/worker_processes", Dir: false, Value: "4"},
			},
		},
	}

	c.AddResponse("/foo", fooResp)
	c.AddResponse("/quux", quuxResp)
	c.AddResponse("/nginx", nginxResp)
	keys := []string{"/nginx", "/foo", "/quux"}
	values, err := GetValues(c, "", keys)
	if err != nil {
		t.Error(err.Error())
	}
	if values["nginx_port"] != "443" {
		t.Errorf("Expected nginx_port to == 443, got %s", values["nginx_port"])
	}
	if values["nginx_worker_processes"] != "4" {
		t.Errorf("Expected nginx_worker_processes == 4, got %s", values["nginx_worker_processes"])
	}
	if values["foo_three_bar"] != "three_bar" {
		t.Errorf("Expected foo_three_bar == three_bar, got %s", values["foo_three_bar"])
	}
	if values["quux"] != "foo" {
		t.Errorf("Expected quux == foo, got %s", values["quux"])
	}
}
