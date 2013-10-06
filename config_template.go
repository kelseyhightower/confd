// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"encoding/json"
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"github.com/kelseyhightower/confd/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type ConfigTemplate struct {
	Path      string
	Templates []Template
	Services  map[string]Service
}

type Service struct {
	Name string
	Cmd  string
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

func NewConfigTemplateFromFile(path string) (*ConfigTemplate, error) {
	var ct *ConfigTemplate
	ct.Path = path
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return ct, err
	}
	if err = json.Unmarshal(f, &ct); err != nil {
		return ct, err
	}
	return ct, nil
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

func (ct *ConfigTemplate) Process() error {
	tmplDir := config.Confd.TemplateDir
	for _, t := range ct.Templates {
		if err := t.GetValuesFromEctd(); err != nil {
			return err
		}
		src := filepath.Join(tmplDir, t.Src)
		if IsFileExist(src) {
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
			err, ok := SameFile(temp.Name(), t.Dest)
			if err != nil {
				log.Error(err.Error())
			}
			if !ok {
				log.Info(t.Dest + " not in sync")
				os.Rename(temp.Name(), t.Dest)
				cmd := ct.Services[t.Service].Cmd
				log.Info("Running " + cmd)
			}
		} else {
			return errors.New("Missing template: " + src)
		}
	}
	return nil
}
