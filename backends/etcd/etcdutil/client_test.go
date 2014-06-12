// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package etcdutil

import (
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/confd/backends/etcd/etcdtest"
)

func TestGetValues(t *testing.T) {
	// Use stub etcd client.
	c := etcdtest.NewClient()
	client := &Client{c}

	fooResp := &etcd.Response{
		Action: "GET",
		Node: &etcd.Node{
			Key:   "/foo",
			Dir:   true,
			Value: "",
			Nodes: etcd.Nodes{
				&etcd.Node{Key: "/foo/one", Dir: false, Value: "one"},
				&etcd.Node{Key: "foo/two", Dir: false, Value: "two"},
				&etcd.Node{
					Key:   "/foo/three",
					Dir:   true,
					Value: "",
					Nodes: etcd.Nodes{
						&etcd.Node{Key: "/foo/three/bar", Value: "three_bar", Dir: false},
					},
				},
			},
		},
	}
	quuxResp := &etcd.Response{
		Action: "GET",
		Node:   &etcd.Node{Key: "/quux", Dir: false, Value: "foo"},
	}
	nginxResp := &etcd.Response{
		Action: "GET",
		Node: &etcd.Node{
			Key:   "/nginx",
			Value: "",
			Dir:   true,
			Nodes: etcd.Nodes{
				&etcd.Node{Key: "/nginx/port", Dir: false, Value: "443"},
				&etcd.Node{Key: "/nginx/worker_processes", Dir: false, Value: "4"},
			},
		},
	}

	c.AddResponse("/foo", fooResp)
	c.AddResponse("/quux", quuxResp)
	c.AddResponse("/nginx", nginxResp)
	keys := []string{"/nginx", "/foo", "/quux"}
	values, err := client.GetValues(keys)
	if err != nil {
		t.Error(err.Error())
	}
	if values["/nginx/port"] != "443" {
		t.Errorf("Expected nginx_port to == 443, got %s", values["nginx_port"])
	}
	if values["/nginx/worker_processes"] != "4" {
		t.Errorf("Expected nginx_worker_processes == 4, got %s", values["nginx_worker_processes"])
	}
	if values["/foo/three/bar"] != "three_bar" {
		t.Errorf("Expected foo_three_bar == three_bar, got %s", values["foo_three_bar"])
	}
	if values["/quux"] != "foo" {
		t.Errorf("Expected quux == foo, got %s", values["quux"])
	}
}
