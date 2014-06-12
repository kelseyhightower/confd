// Copyright (c) 2014 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package template

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/confd/log"
)

// fileInfo describes a configuration file and is returned by fileStat.
type fileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode os.FileMode
	Md5  string
}

var replacer = strings.NewReplacer("/", "_", "-", "_")

func appendPrefix(prefix string, keys []string) []string {
	s := make([]string, len(keys))
	for i, k := range keys {
		s[i] = path.Join(prefix, k)
	}
	return s
}

// cleanKeys is used to transform the path based keys we
// get from the StoreClient to a more friendly format.
func cleanKeys(vars map[string]interface{}, prefix string) map[string]interface{} {
	clean := make(map[string]interface{}, len(vars))
	for key, val := range vars {
		clean[pathToKey(key, prefix)] = val
	}
	return clean
}

func mapNodes(node *etcd.Node) map[string]interface{} {
	result := make(map[string]interface{})
	for _, node := range node.Nodes {
		if len(node.Nodes) > 0 {
			result[node.Key] = node.Nodes
		} else {
			result[node.Key] = node.Value
		}
	}
	return cleanKeys(result, node.Key)
}

// isFileExist reports whether path exits.
func isFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}

// pathToKey translates etcd key paths into something more suitable for use
// in Golang templates. Turn /prefix/key/subkey into key_subkey.
func pathToKey(key, prefix string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	key = strings.TrimPrefix(key, "/")
	key = strings.TrimPrefix(key, prefix)
	key = strings.TrimPrefix(key, "/")
	return replacer.Replace(key)
}

// fileStat return a fileInfo describing the named file.
func fileStat(name string) (fi fileInfo, err error) {
	if isFileExist(name) {
		f, err := os.Open(name)
		if err != nil {
			return fi, err
		}
		defer f.Close()
		stats, _ := f.Stat()
		fi.Uid = stats.Sys().(*syscall.Stat_t).Uid
		fi.Gid = stats.Sys().(*syscall.Stat_t).Gid
		fi.Mode = stats.Mode()
		h := md5.New()
		io.Copy(h, f)
		fi.Md5 = fmt.Sprintf("%x", h.Sum(nil))
		return fi, nil
	} else {
		return fi, errors.New("File not found")
	}
}

// sameConfig reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func sameConfig(src, dest string) (bool, error) {
	if !isFileExist(dest) {
		return false, nil
	}
	d, err := fileStat(dest)
	if err != nil {
		return false, err
	}
	s, err := fileStat(src)
	if err != nil {
		return false, err
	}
	if d.Uid != s.Uid {
		log.Info(fmt.Sprintf("%s has UID %d should be %d", dest, d.Uid, s.Uid))
	}
	if d.Gid != s.Gid {
		log.Info(fmt.Sprintf("%s has GID %d should be %d", dest, d.Gid, s.Gid))
	}
	if d.Mode != s.Mode {
		log.Info(fmt.Sprintf("%s has mode %s should be %s", dest, os.FileMode(d.Mode), os.FileMode(s.Mode)))
	}
	if d.Md5 != s.Md5 {
		log.Info(fmt.Sprintf("%s has md5sum %s should be %s", dest, d.Md5, s.Md5))
	}
	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Md5 != s.Md5 {
		return false, nil
	}
	return true, nil
}
