package memkv

import (
	"path/filepath"
	"reflect"
	"testing"
)

var gettests = []struct {
	key   string
	value string
	err   error
	want  string
}{
	{"/db/user", "admin", nil, "admin"},
	{"/db/pass", "foo", nil, "foo"},
	{"/missing", "", ErrNotExist, ""},
}

func TestGet(t *testing.T) {
	for _, tt := range gettests {
		s := New()
		if tt.err == nil {
			s.Set(tt.key, tt.value)
		}
		got, err := s.Get(tt.key)
		if got != tt.want || err != tt.err {
			t.Errorf("Get(%q) = %v, %v, want %v, %v", tt.key, got, err, tt.want, tt.err)
		}
	}
}

var globtestinput = map[string]string{
	"/app/db/pass":               "foo",
	"/app/db/user":               "admin",
	"/app/port":                  "443",
	"/app/url":                   "app.example.com",
	"/app/vhosts/host1":          "app.example.com",
	"/app/upstream/host1":        "203.0.113.0.1:8080",
	"/app/upstream/host1/domain": "app.example.com",
	"/app/upstream/host2":        "203.0.113.0.2:8080",
	"/app/upstream/host2/domain": "app.example.com",
}

var globtests = []struct {
	pattern string
	err     error
	want    []Node
}{
	{"/app/db/*", nil,
		[]Node{
			Node{"/app/db/pass", "foo"},
			Node{"/app/db/user", "admin"}}},
	{"/app/*/host1", nil,
		[]Node{
			Node{"/app/upstream/host1", "203.0.113.0.1:8080"},
			Node{"/app/vhosts/host1", "app.example.com"}}},

	{"/app/upstream/*", nil,
		[]Node{
			Node{"/app/upstream/host1", "203.0.113.0.1:8080"},
			Node{"/app/upstream/host2", "203.0.113.0.2:8080"}}},
	{"[]a]", filepath.ErrBadPattern, nil},
}

func TestGlob(t *testing.T) {
	s := New()
	for k, v := range globtestinput {
		s.Set(k, v)
	}
	for _, tt := range globtests {
		got, err := s.Glob(tt.pattern)
		if !reflect.DeepEqual([]Node(got), []Node(tt.want)) || err != tt.err {
			t.Errorf("Glob(%q) = %v, %v, want %v, %v", tt.pattern, got, err, tt.want, tt.err)
		}
	}
}

func TestDelete(t *testing.T) {
	s := New()
	s.Set("/app/port", "8080")
	want := "8080"
	got, err := s.Get("/app/port")
	if err != nil || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, true)
	}
	// Delete the node.
	s.Delete("/app/port")
	want = ""
	got, err = s.Get("/app/port")
	if err != ErrNotExist || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, false)
	}
	// Delete a missing node.
	s.Delete("/app/port")
}
