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

func ProcessConfig(config string) error {
	f, err := ioutil.ReadFile(config)
	if err != nil {
		return err
	}

	var cfg *Config
	if err = json.Unmarshal(f, &cfg); err != nil {
		return err
	}

	for _, tmplConfig := range cfg.Templates {
		m, err := GetValuesFromEctd(tmplConfig.Keys)
		if err != nil {
			return err
		}
		tmpl := filepath.Join(settings.ConfigDir, "templates", tmplConfig.Src)
		if isFileExist(tmpl) {
			temp, err := ioutil.TempFile("", "")
			defer os.Remove(temp.Name())
			if err != nil {
				return err
			}

			data, err := ioutil.ReadFile(tmpl)
			if err != nil {
				return err
			}
			t := template.Must(template.New("test").Parse(string(data)))
			err = t.Execute(temp, m)
			if err != nil {
				return err
			}
			myMode, _ := strconv.ParseUint(tmplConfig.Mode, 0, 32)
			os.Chmod(temp.Name(), os.FileMode(myMode))
			os.Chown(temp.Name(), tmplConfig.Uid, tmplConfig.Gid)

			if isSync(temp.Name(), tmplConfig.Dest) {
				log.Print("Files are in sync")
			} else {
				log.Print("File not in sync")
				os.Rename(temp.Name(), tmplConfig.Dest)
				cmd := cfg.Services[tmplConfig.Service].ReloadCmd
				log.Printf("Running %s", cmd)
			}
		} else {
			return errors.New("Missing template: " + tmpl)
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
	// Set defaults
	settings.ConfigDir = "/etc/confd/conf.d"
	settings.EtcdURL = "http://0.0.0.0:4001"
	settings.EtcdPrefix = "/"

	// Override defaults from config file.
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
