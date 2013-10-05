// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/confd/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"
)

var (
	config     Config
	configFile = "/etc/confd/confd.toml"
)

type Config struct {
	Confd confdConfig
	Etcd  etcdConfig
}

type confdConfig struct {
	ConfigDir   string
	TemplateDir string
	Interval    int
}

type etcdConfig struct {
	Prefix   string
	Machines []string
}

type ConfigTemplate struct {
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
	Vars    map[string]interface{}
}

type FileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode uint32
	Md5  string
}

func main() {
	if err := setConfig(); err != nil {
		log.Fatal(err.Error())
	}
	configTemplates, err := filepath.Glob(filepath.Join(config.Confd.ConfigDir, "*json"))
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		for _, ct := range configTemplates {
			if err := ProcessConfigTemplate(ct); err != nil {
				log.Error(err.Error())
			}
		}
		interval := config.Confd.Interval
		if err != nil {
			log.Fatal(err.Error())
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func NewConfigTemplateFromFile(name string) (*ConfigTemplate, error) {
	var ct *ConfigTemplate
	f, err := ioutil.ReadFile(name)
	if err != nil {
		return ct, err
	}
	if err = json.Unmarshal(f, &ct); err != nil {
		return ct, err
	}
	return ct, nil
}

func ProcessConfigTemplate(configTemplate string) error {
	tmplDir := config.Confd.TemplateDir
	ct, err := NewConfigTemplateFromFile(configTemplate)
	if err != nil {
		return err
	}
	for _, t := range ct.Templates {
		if err := t.GetValuesFromEctd(); err != nil {
			return err
		}
		src := filepath.Join(tmplDir, t.Src)
		if isFileExist(src) {
			temp, err := ioutil.TempFile("", "")
			defer os.Remove(temp.Name())
			if err != nil {
				return err
			}

			tmpl := template.Must(template.New(t.Src).ParseFiles(src))
			if err = tmpl.Execute(temp, t.Vars); err != nil {
				return err
			}
			if err = t.SetFileAttrs(temp.Name()); err != nil {
				return err
			}
			if !isSync(temp.Name(), t.Dest) {
				log.Info(t.Dest + " not in sync")
				os.Rename(temp.Name(), t.Dest)
				cmd := ct.Services[t.Service].ReloadCmd
				log.Info("Running " + cmd)
			}
		} else {
			return errors.New("Missing template: " + src)
		}
	}
	return nil
}

func (t *Template) GetValuesFromEctd() error {
	var (
		prefix   = config.Etcd.Prefix
		machines = config.Etcd.Machines
	)

	c := etcd.NewClient()
	success := c.SetCluster(machines)
	if !success {
		log.Fatal("could not sync machines")
	}
	r := strings.NewReplacer("/", "_")
	t.Vars = make(map[string]interface{})
	for _, key := range t.Keys {
		values, err := c.Get(filepath.Join(prefix, key))
		if err != nil {
			return err
		}
		for _, v := range values {
			key := strings.TrimPrefix(v.Key, prefix)
			new_key := r.Replace(key)
			t.Vars[new_key] = v.Value
		}
	}
	return nil
}

func (t *Template) SetFileAttrs(name string) error {
	mode, _ := strconv.ParseUint(t.Mode, 0, 32)
	os.Chmod(name, os.FileMode(mode))
	os.Chown(name, t.Uid, t.Gid)
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
	etcdDefaults := etcdConfig{
		Prefix:   "/",
		Machines: []string{"http://127.0.0.1:4001"},
	}
	confdDefaults := confdConfig{
		Interval:    600,
		ConfigDir:   "/etc/confd/conf.d",
		TemplateDir: "/etc/confd/templates",
	}
	config.Etcd = etcdDefaults
	config.Confd = confdDefaults

	if isFileExist(configFile) {
		_, err := toml.DecodeFile(configFile, &config)
		if err != nil {
			return err
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
