package main

import (
	"testing"
)

var pathToKeyTests = []struct {
	key      string
	prefix   string
	expected string
}{
	{"/nginx/port", "", "nginx_port"},
	{"/prefix/nginx/port", "/prefix", "nginx_port"},
	{"/nginx/worker_processes", "", "nginx_worker_processes"},
	{"/foo/bar/mat/zoo", "", "foo_bar_mat_zoo"},
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
