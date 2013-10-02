// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/go-ini"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"
)

type FileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode uint16
	Md5  string
}

type Settings struct {
	ConfigDir  string
	EtcdURL    string
	EtcdPrefix string
}

type Config struct {
	Templates []Template
	Services  map[string]Service
}

type Service struct {
	Name      string
	ReloadCmd string `json:"reload_cmd"`
}

type Template struct {
	Dest    string
	Gid     int
	Keys    []string
	Mode    string
	Uid     int
	Service string
	Src     string
}

var settings Settings
var defaultConfig = "/etc/confd/confd.ini"

func main() {
	if err := setConfig(); err != nil {
		log.Fatal(err.Error())
	}
	configs, err := filepath.Glob(filepath.Join(settings.ConfigDir, "*json"))
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, config := range configs {
		if err := ProcessConfig(config); err != nil {
			log.Println(err.Error())
		}
	}
}

func GetValuesFromEctd(keys []string) (map[string]interface{}, error) {
	c := etcd.NewClient()
	r := strings.NewReplacer("/", "_")
	m := make(map[string]interface{})
	for _, key := range keys {
		values, err := c.Get(filepath.Join(settings.EtcdPrefix, key))
		if err != nil {
			return m, err
		}
		for _, v := range values {
			key := strings.TrimPrefix(v.Key, settings.EtcdPrefix)
			new_key := r.Replace(key)
			m[new_key] = v.Value
		}
	}
	return m, nil
}

func NewConfigFromFile(name string) (*Config, error) {
	var c *Config
	f, err := ioutil.ReadFile(name)
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(f, &c); err != nil {
		return c, err
	}
	return c, nil
}

func ProcessConfig(config string) error {
	c, err := NewConfigFromFile(config)
	if err != nil {
		return err
	}
	for _, t := range c.Templates {
		m, err := GetValuesFromEctd(t.Keys)
		if err != nil {
			return err
		}
		src := filepath.Join(settings.ConfigDir, "templates", t.Src)
		if isFileExist(src) {
			temp, err := ioutil.TempFile("", "")
			defer os.Remove(temp.Name())
			if err != nil {
				return err
			}

			tmpl := template.Must(template.New(t.Src).ParseFiles(src))
			err = tmpl.Execute(temp, m)
			if err != nil {
				return err
			}
			mode, _ := strconv.ParseUint(t.Mode, 0, 32)
			os.Chmod(temp.Name(), os.FileMode(mode))
			os.Chown(temp.Name(), t.Uid, t.Gid)

			if isSync(temp.Name(), t.Dest) {
				log.Print("Files are in sync")
			} else {
				log.Print("File not in sync")
				os.Rename(temp.Name(), t.Dest)
				cmd := c.Services[t.Service].ReloadCmd
				log.Printf("Running %s", cmd)
			}
		} else {
			return errors.New("Missing template: " + src)
		}
	}
	return nil
}

func Stat(name string) (fi FileInfo, err error) {
	if isFileExist(name) {
		f, err := os.Open(name)
		defer f.Close()
		if err != nil {
			return fi, err
		}
		stats, _ := f.Stat()
		fi.Uid = stats.Sys().(*syscall.Stat_t).Uid
		fi.Gid = stats.Sys().(*syscall.Stat_t).Gid
		fi.Mode = stats.Sys().(*syscall.Stat_t).Mode
		h := md5.New()
		io.Copy(h, f)
		fi.Md5 = fmt.Sprintf("%x", h.Sum(nil))
		return fi, nil
	} else {
		return fi, errors.New("File not found")
	}
}

func isSync(src, dest string) bool {
	if !isFileExist(dest) {
		return false
	}
	old, err := Stat(dest)
	if err != nil {
		log.Fatal(err.Error())
	}

	n, err := Stat(src)
	if err != nil {
		log.Fatal(err.Error())
	}
	if old.Uid != n.Uid {
		return false
	}
	if old.Gid != n.Gid {
		return false
	}
	if old.Mode != n.Mode {
		return false
	}
	if old.Md5 != n.Md5 {
		return false
	}
	return true
}

func setConfig() error {
	settings.ConfigDir = "/etc/confd/conf.d"
	settings.EtcdURL = "http://0.0.0.0:4001"
	settings.EtcdPrefix = "/"

	if isFileExist(defaultConfig) {
		s, err := ini.LoadFile(defaultConfig)
		if err != nil {
			return err
		}
		if configDir, ok := s.Get("main", "config_dir"); ok {
			settings.ConfigDir = configDir
		}
		if etcdURL, ok := s.Get("etcd", "url"); ok {
			settings.EtcdURL = etcdURL
		}
		if etcdPrefix, ok := s.Get("etcd", "prefix"); ok {
			settings.EtcdPrefix = etcdPrefix
		}
	}
	return nil
}

func isFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
