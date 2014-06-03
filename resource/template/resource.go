// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package template

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/node"
)

var replacer = strings.NewReplacer("/", "_")

// TemplateResourceConfig holds the parsed template resource.
type TemplateResourceConfig struct {
	TemplateResource TemplateResource `toml:"template"`
}

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	Dest        string
	FileMode    os.FileMode
	Gid         int
	Keys        []string
	Mode        string
	Uid         int
	ReloadCmd   string `toml:"reload_cmd"`
	CheckCmd    string `toml:"check_cmd"`
	StageFile   *os.File
	Src         string
	Vars        map[string]interface{}
	Dirs        node.Directory
	storeClient StoreClient
}

// NewTemplateResourceFromPath creates a TemplateResource using a decoded file path
// and the supplied StoreClient as input.
// It returns a TemplateResource and an error if any.
func NewTemplateResourceFromPath(path string, s StoreClient) (*TemplateResource, error) {
	if s == nil {
		return nil, errors.New("A valid StoreClient is required.")
	}
	var tc *TemplateResourceConfig
	log.Debug("Loading template resource from " + path)
	_, err := toml.DecodeFile(path, &tc)
	if err != nil {
		return nil, fmt.Errorf("Cannot process template resource %s - %s", path, err.Error())
	}
	tc.TemplateResource.storeClient = s
	return &tc.TemplateResource, nil
}

// setVars sets the Vars for template resource.
func (t *TemplateResource) setVars() error {
	var err error
	log.Debug("Retrieving keys from store")
	log.Debug("Key prefix set to " + config.Prefix())
	vars, err := t.storeClient.GetValues(appendPrefix(config.Prefix(), t.Keys))
	if err != nil {
		return err
	}
	t.setDirs(vars)
	t.Vars = cleanKeys(vars, config.Prefix())
	return nil
}

// setDirs sets the Dirs for the template resource.
// All keys are grouped based on their directory path names.
// For example, /upstream/app1 and upstream/app2 will be grouped as
//	{
//		"upstream": []Node{
//			{"app1": value}},
//			{"app2": value}},
//		 }
//	}
//
// Dirs are exposed to resource templated to enable iteration.
func (t *TemplateResource) setDirs(vars map[string]interface{}) {
	d := node.NewDirectory()
	for k, v := range vars {
		directory := filepath.Dir(filepath.Join("/", strings.TrimPrefix(k, config.Prefix())))
		d.Add(pathToKey(directory, config.Prefix()), node.Node{filepath.Base(k), v})
	}
	t.Dirs = d
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (t *TemplateResource) createStageFile() error {
	t.Src = filepath.Join(config.TemplateDir(), t.Src)
	log.Debug("Using source template " + t.Src)
	if !isFileExist(t.Src) {
		return errors.New("Missing template: " + t.Src)
	}
	// create TempFile in Dest directory to avoid cross-filesystem issues
	temp, err := ioutil.TempFile(filepath.Dir(t.Dest), "."+filepath.Base(t.Dest))
	if err != nil {
		os.Remove(temp.Name())
		return err
	}
	defer temp.Close()
	log.Debug("Compiling source template " + t.Src)
	tplFuncMap := make(template.FuncMap)
	tplFuncMap["Base"] = path.Base

	tplFuncMap["GetDir"] = t.Dirs.Get
	tplFuncMap["MapDir"] = mapNodes
	tplFuncMap["GetEnv"] = os.Getenv
	tmpl := template.Must(template.New(path.Base(t.Src)).Funcs(tplFuncMap).ParseFiles(t.Src))
	if err = tmpl.Execute(temp, t.Vars); err != nil {
		return err
	}
	// Set the owner, group, and mode on the stage file now to make it easier to
	// compare against the destination configuration file later.
	os.Chmod(temp.Name(), t.FileMode)
	os.Chown(temp.Name(), t.Uid, t.Gid)
	t.StageFile = temp
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *TemplateResource) sync() error {
	staged := t.StageFile.Name()
	defer os.Remove(staged)
	log.Debug("Comparing candidate config to " + t.Dest)
	ok, err := sameConfig(staged, t.Dest)
	if err != nil {
		log.Error(err.Error())
	}
	if config.Noop() {
		log.Warning("Noop mode enabled " + t.Dest + " will not be modified")
		return nil
	}
	if !ok {
		log.Info("Target config " + t.Dest + " out of sync")
		if t.CheckCmd != "" {
			if err := t.check(); err != nil {
				return errors.New("Config check failed: " + err.Error())
			}
		}
		log.Debug("Overwriting target config " + t.Dest)
		if err := os.Rename(staged, t.Dest); err != nil {
			return err
		}
		if t.ReloadCmd != "" {
			if err := t.reload(); err != nil {
				return err
			}
		}
		log.Info("Target config " + t.Dest + " has been updated")
	} else {
		log.Info("Target config " + t.Dest + " in sync")
	}
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (t *TemplateResource) check() error {
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = t.StageFile.Name()
	tmpl, err := template.New("checkcmd").Parse(t.CheckCmd)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&cmdBuffer, data); err != nil {
		return err
	}
	log.Debug("Running " + cmdBuffer.String())
	c := exec.Command("/bin/sh", "-c", cmdBuffer.String())
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (t *TemplateResource) reload() error {
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["dest"] = t.Dest
	tmpl, err := template.New("reloadcmd").Parse(t.ReloadCmd)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&cmdBuffer, data); err != nil {
		return err
	}
	log.Debug("Running " + cmdBuffer.String())
	c := exec.Command("/bin/sh", "-c", cmdBuffer.String())
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

// process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *TemplateResource) process() error {
	if err := t.setFileMode(); err != nil {
		return err
	}
	if err := t.setVars(); err != nil {
		return err
	}
	if err := t.createStageFile(); err != nil {
		return err
	}
	if err := t.sync(); err != nil {
		return err
	}
	return nil
}

// setFileMode sets the FileMode.
func (t *TemplateResource) setFileMode() error {
	if t.Mode == "" {
		if !isFileExist(t.Dest) {
			t.FileMode = 0644
		} else {
			fi, err := os.Stat(t.Dest)
			if err != nil {
				return err
			}
			t.FileMode = fi.Mode()
		}
	} else {
		mode, err := strconv.ParseUint(t.Mode, 0, 32)
		if err != nil {
			return err
		}
		t.FileMode = os.FileMode(mode)
	}
	return nil
}

// ProcessTemplateResources is a convenience function that loads all the
// template resources and processes them serially. Called from main.
// It returns a list of errors if any.
func ProcessTemplateResources(s StoreClient) []error {
	runErrors := make([]error, 0)
	var err error
	if s == nil {
		runErrors = append(runErrors, errors.New("A StoreClient client is required"))
		return runErrors
	}
	log.Debug("Loading template resources from confdir " + config.ConfDir())
	if !isFileExist(config.ConfDir()) {
		log.Warning(fmt.Sprintf("Cannot load template resources confdir '%s' does not exist", config.ConfDir()))
		return runErrors
	}
	paths, err := filepath.Glob(filepath.Join(config.ConfigDir(), "*toml"))
	if err != nil {
		runErrors = append(runErrors, err)
		return runErrors
	}
	for _, p := range paths {
		log.Debug("Processing template resource " + p)
		tplFuncMap := make(template.FuncMap)
		tplFuncMap["Base"] = path.Base
		tplFuncMap["GetEnv"] = os.Getenv
		data := make(map[string]string)
		data["file"] = path.Base(p)
		data["src"] = "{{ .src }}"
		data["dest"] = "{{ .dest }}"
		resTemp, terr := template.New(path.Base(p)).Funcs(tplFuncMap).ParseFiles(p)
		if terr != nil {
			panic(terr)
		}
		var doc bytes.Buffer
		resTemp.Execute(&doc, data)
		docs := doc.String()
		var tc *TemplateResourceConfig
		log.Debug(fmt.Sprintf("Resource parsed as: %s", docs))
		_, err := toml.Decode(docs, &tc)
		if err != nil {
			runErrors = append(runErrors, err)
			log.Error(err.Error())
			continue
		}
		tc.TemplateResource.storeClient = s
		if err := tc.TemplateResource.process(); err != nil {
			runErrors = append(runErrors, err)
			log.Error(err.Error())
			continue
		}
		log.Debug("Processing of template resource " + p + " complete")
	}
	return runErrors
}
