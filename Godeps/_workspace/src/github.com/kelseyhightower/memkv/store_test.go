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
	want  KVPair
}{
	{"/db/user", "admin", nil, KVPair{"/db/user", "admin"}},
	{"/db/pass", "foo", nil, KVPair{"/db/pass", "foo"}},
	{"/missing", "", ErrNotExist, KVPair{}},
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

var getalltestinput = map[string]string{
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

var getalltests = []struct {
	pattern string
	err     error
	want    []KVPair
}{
	{"/app/db/*", nil,
		[]KVPair{
			KVPair{"/app/db/pass", "foo"},
			KVPair{"/app/db/user", "admin"}}},
	{"/app/*/host1", nil,
		[]KVPair{
			KVPair{"/app/upstream/host1", "203.0.113.0.1:8080"},
			KVPair{"/app/vhosts/host1", "app.example.com"}}},

	{"/app/upstream/*", nil,
		[]KVPair{
			KVPair{"/app/upstream/host1", "203.0.113.0.1:8080"},
			KVPair{"/app/upstream/host2", "203.0.113.0.2:8080"}}},
	{"[]a]", filepath.ErrBadPattern, nil},
}

func TestGetAll(t *testing.T) {
	s := New()
	for k, v := range getalltestinput {
		s.Set(k, v)
	}
	for _, tt := range getalltests {
		got, err := s.GetAll(tt.pattern)
		if !reflect.DeepEqual([]KVPair(got), []KVPair(tt.want)) || err != tt.err {
			t.Errorf("GetAll(%q) = %v, %v, want %v, %v", tt.pattern, got, err, tt.want, tt.err)
		}
	}
}

func TestDel(t *testing.T) {
	s := New()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got, err := s.Get("/app/port")
	if err != nil || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, true)
	}
	s.Del("/app/port")
	want = KVPair{}
	got, err = s.Get("/app/port")
	if err != ErrNotExist || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, false)
	}
	s.Del("/app/port")
}

var listTestMap = map[string]string{
	"/deis/database/user":            "user",
	"/deis/database/pass":            "pass",
	"/deis/services/key":             "value",
	"/deis/services/notaservice/foo": "bar",
	"/deis/services/srv1/node1":      "10.244.1.1:80",
	"/deis/services/srv1/node2":      "10.244.1.2:80",
	"/deis/services/srv1/node3":      "10.244.1.3:80",
	"/deis/services/srv2/node1":      "10.244.2.1:80",
	"/deis/services/srv2/node2":      "10.244.2.2:80",
}

func TestList(t *testing.T) {
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	want := []string{"key", "notaservice", "srv1", "srv2"}
	got := s.List("/deis/services")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List(%s) = %v, want %v", "/deis/services", got, want)
	}
}

func TestListDir(t *testing.T) {
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	want := []string{"notaservice", "srv1", "srv2"}
	got := s.ListDir("/deis/services")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List(%s) = %v, want %v", "/deis/services", got, want)
	}
}
