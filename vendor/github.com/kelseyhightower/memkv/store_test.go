package memkv

import (
	"path"
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
	{"/missing", "", &KeyError{"/missing", ErrNotExist}, KVPair{}},
}

func TestGet(t *testing.T) {
	for _, tt := range gettests {
		s := New()
		if tt.err == nil {
			s.Set(tt.key, tt.value)
		}
		got, err := s.Get(tt.key)
		if got != tt.want || !reflect.DeepEqual(err, tt.err) {
			t.Errorf("Get(%q) = %v, %v, want %v, %v", tt.key, got, err, tt.want, tt.err)
		}
	}
}

var getvtests = []struct {
	key   string
	value string
	err   error
	want  string
}{
	{"/db/user", "admin", nil, "admin"},
	{"/db/pass", "foo", nil, "foo"},
	{"/missing", "", &KeyError{"/missing", ErrNotExist}, ""},
}

func TestGetValue(t *testing.T) {
	for _, tt := range getvtests {
		s := New()
		if tt.err == nil {
			s.Set(tt.key, tt.value)
		}
		got, err := s.GetValue(tt.key)
		if got != tt.want || !reflect.DeepEqual(err, tt.err) {
			t.Errorf("Get(%q) = %v, %v, want %v, %v", tt.key, got, err, tt.want, tt.err)
		}
	}
}

func TestGetValueWithDefault(t *testing.T) {
	want := "defaultValue"
	s := New()
	got, err := s.GetValue("/db/user", "defaultValue")
	if err != nil {
		t.Errorf("Unexpected error", err.Error())
	}
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestGetValueWithEmptyDefault(t *testing.T) {
	want := ""
	s := New()
	got, err := s.GetValue("/db/user", "")
	if err != nil {
		t.Errorf("Unexpected error", err.Error())
	}
	if got != want {
		t.Errorf("want %v, got %v", want, got)
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
	{"[]a]", path.ErrBadPattern, nil},
	{"/app/missing/*", nil, []KVPair{}},
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
	if !reflect.DeepEqual(err, &KeyError{"/app/port", ErrNotExist}) || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, false)
	}
	s.Del("/app/port")
}

func TestPurge(t *testing.T) {
	s := New()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got, err := s.Get("/app/port")
	if err != nil || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, true)
	}
	s.Purge()
	want = KVPair{}
	got, err = s.Get("/app/port")
	if !reflect.DeepEqual(err, &KeyError{"/app/port", ErrNotExist}) || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, false)
	}
	s.Set("/app/port", "8080")
	want = KVPair{"/app/port", "8080"}
	got, err = s.Get("/app/port")
	if err != nil || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, err, want, true)
	}
}

var listTestMap = map[string]string{
	"/deis/database/user":             "user",
	"/deis/database/pass":             "pass",
	"/deis/services/key":              "value",
	"/deis/services/notaservice/foo":  "bar",
	"/deis/services/srv1/node1":       "10.244.1.1:80",
	"/deis/services/srv1/node2":       "10.244.1.2:80",
	"/deis/services/srv1/node3":       "10.244.1.3:80",
	"/deis/services/srv2/node1":       "10.244.2.1:80",
	"/deis/services/srv2/node2":       "10.244.2.2:80",
	"/deis/prefix/node1":              "prefix_node1",
	"/deis/prefix/node2/leafnode":     "prefix_node2",
	"/deis/prefix/node3/leafnode":     "prefix_node3",
	"/deis/prefix_a/node4":            "prefix_a_node4",
	"/deis/prefixb/node5/leafnode":    "prefixb_node5",
	"/deis/dirprefix/node1":           "prefix_node1",
	"/deis/dirprefix/node2/leafnode":  "prefix_node2",
	"/deis/dirprefix/node3/leafnode":  "prefix_node3",
	"/deis/dirprefix_a/node4":         "prefix_a_node4",
	"/deis/dirprefixb/node5/leafnode": "prefixb_node5",
}

func TestList(t *testing.T) {
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	want := []string{"key", "notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	for _, filePath := range paths {
		got := s.List(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestListForSamePrefix(t *testing.T) {
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	want := []string{"node1", "node2", "node3"}
	paths := []string{
		"/deis/prefix",
		"/deis/prefix/",
	}
	for _, filePath := range paths {
		got := s.List(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestListForFile(t *testing.T) {
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	want := []string{"key"}
	got := s.List("/deis/services/key")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List(%s) = %v, want %v", "/deis/services", got, want)
	}
}

func TestListEmptyChildrenTrailingSlash(t *testing.T) {
	s := New()
	s.Set("/top/first", "")
	s.Set("/top/second", "")

	want := []string{}
	paths := []string{"/top/first/", "/top/second/"}
	for _, filePath := range paths {
		got := s.List(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestListEmptyChildren(t *testing.T) {
	s := New()
	s.Set("/top/first", "")
	s.Set("/top/second", "")

	first := s.List("/top/first")
	if !reflect.DeepEqual(first, []string{"first"}) {
		t.Errorf("List(/top/first) = %v, want [first]", first)
	}

	second := s.List("/top/second")
	if !reflect.DeepEqual(second, []string{"second"}) {
		t.Errorf("List(/top/second) = %v, want [second]", second)
	}
}

func TestListDir(t *testing.T) {
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	want := []string{"notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	for _, filePath := range paths {
		got := s.ListDir(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestListDirForSamePrefix(t *testing.T) {
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	want := []string{"node2", "node3"}
	paths := []string{
		"/deis/dirprefix",
		"/deis/dirprefix/",
	}
	for _, filePath := range paths {
		got := s.ListDir(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}
