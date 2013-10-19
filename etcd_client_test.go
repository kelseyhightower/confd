package main

import (
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
