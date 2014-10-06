// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package template

import (
	"bytes"
	"encoding/json"
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
	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/memkv"
)

type Config struct {
	ConfDir       string
	ConfigDir     string
	KeepStageFile bool
	Noop          bool
	Prefix        string
	StoreClient   backends.StoreClient
	TemplateDir   string
}

// TemplateResourceConfig holds the parsed template resource.
type TemplateResourceConfig struct {
	TemplateResource TemplateResource `toml:"template"`
}

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	CheckCmd      string `toml:"check_cmd"`
	Dest          string
	FileMode      os.FileMode
	Gid           int
	Keys          []string
	Mode          string
	Prefix        string
	ReloadCmd     string `toml:"reload_cmd"`
	Src           string
	StageFile     *os.File
	Uid           int
	keepStageFile bool
	noop          bool
	prefix        string
	store         memkv.Store
	storeClient   backends.StoreClient
}

var ErrEmptySrc = errors.New("empty src template")

// New creates a TemplateResource.
func New(path string, config Config) (*TemplateResource, error) {
	if config.StoreClient == nil {
		return nil, errors.New("A valid StoreClient is required.")
	}
	var tc *TemplateResourceConfig
	log.Debug("Loading template resource from " + path)
	_, err := toml.DecodeFile(path, &tc)
	if err != nil {
		return nil, fmt.Errorf("Cannot process template resource %s - %s", path, err.Error())
	}
	tr := tc.TemplateResource
	tr.keepStageFile = config.KeepStageFile
	tr.noop = config.Noop
	tr.storeClient = config.StoreClient
	tr.store = memkv.New()
	tr.prefix = filepath.Join("/", config.Prefix, tr.Prefix)
	if tr.Src == "" {
		return nil, ErrEmptySrc
	}
	tr.Src = filepath.Join(config.TemplateDir, tr.Src)
	return &tr, nil
}

// setVars sets the Vars for template resource.
func (t *TemplateResource) setVars() error {
	var err error
	log.Debug("Retrieving keys from store")
	log.Debug("Key prefix set to " + t.prefix)
	result, err := t.storeClient.GetValues(appendPrefix(t.prefix, t.Keys))
	if err != nil {
		return err
	}
	for k, v := range result {
		t.store.Set(filepath.Join("/", strings.TrimPrefix(k, t.prefix)), v)
	}
	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (t *TemplateResource) createStageFile() error {
	log.Debug("Using source template " + t.Src)
	if !isFileExist(t.Src) {
		return errors.New("Missing template: " + t.Src)
	}
	// create TempFile in Dest directory to avoid cross-filesystem issues
	temp, err := ioutil.TempFile(filepath.Dir(t.Dest), "."+filepath.Base(t.Dest))
	if err != nil {
		return err
	}
	defer temp.Close()
	log.Debug("Compiling source template " + t.Src)

	// Add template functions
	tplFuncMap := make(template.FuncMap)
	tplFuncMap["base"] = path.Base
	tplFuncMap["ls"] = t.store.List
	tplFuncMap["lsdir"] = t.store.ListDir
	tplFuncMap["get"] = t.store.Get
	tplFuncMap["gets"] = t.store.GetAll
	tplFuncMap["getv"] = t.store.GetValue
	tplFuncMap["getvs"] = t.store.GetAllValues
	tplFuncMap["split"] = strings.Split
	tplFuncMap["json"] = t.UnmarshalJsonObject
	tplFuncMap["jsonArray"] = t.UnmarshalJsonArray
	tplFuncMap["sibling"] = t.GetSibling
	tplFuncMap["parent"] = path.Dir

	tmpl := template.Must(template.New(path.Base(t.Src)).Funcs(tplFuncMap).ParseFiles(t.Src))
	if err = tmpl.Execute(temp, nil); err != nil {
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
	if t.keepStageFile {
		log.Info("Keeping staged file: " + staged)
	} else {
		defer os.Remove(staged)
	}
	log.Debug("Comparing candidate config to " + t.Dest)
	ok, err := sameConfig(staged, t.Dest)
	if err != nil {
		log.Error(err.Error())
	}
	if t.noop {
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
		err := os.Rename(staged, t.Dest)
		if err != nil {
			if strings.Contains(err.Error(), "device or resource busy") {
				log.Debug("Rename failed - target is likely a mount. Trying to write instead")
				// try to open the file and write to it
				var contents []byte
				var rerr error
				contents, rerr = ioutil.ReadFile(staged)
				if rerr != nil {
					return rerr
				}
				err := ioutil.WriteFile(t.Dest, contents, t.FileMode)
				// make sure owner and group match the temp file, in case the file was created with WriteFile
				os.Chown(t.Dest, t.Uid, t.Gid)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		if t.ReloadCmd != "" {
			if err := t.reload(); err != nil {
				return err
			}
		}
		log.Info("Target config " + t.Dest + " has been updated")
	} else {
		log.Debug("Target config " + t.Dest + " in sync")
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
	output, err := c.CombinedOutput()
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (t *TemplateResource) reload() error {
	log.Debug("Running " + t.ReloadCmd)
	c := exec.Command("/bin/sh", "-c", t.ReloadCmd)
	output, err := c.CombinedOutput()
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("%q", string(output)))
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

func (t *TemplateResource) UnmarshalJsonObject(data string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func (t *TemplateResource) UnmarshalJsonArray(data string) ([]interface{}, error) {
	var ret []interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func (t *TemplateResource) GetSibling(origin string, newKey string) (memkv.KVPair, error) {
	return t.store.Get(path.Join("/", path.Dir(origin), newKey))
}

// ProcessTemplateResources is a convenience function that loads all the
// template resources and processes them serially. Called from main.
// It returns a list of errors if any.
func ProcessTemplateResources(config Config) []error {
	runErrors := make([]error, 0)
	var err error
	if config.StoreClient == nil {
		runErrors = append(runErrors, errors.New("A StoreClient client is required"))
		return runErrors
	}
	log.Debug("Loading template resources from confdir " + config.ConfDir)
	if !isFileExist(config.ConfDir) {
		log.Warning(fmt.Sprintf("Cannot load template resources confdir '%s' does not exist", config.ConfDir))
		return runErrors
	}
	paths, err := filepath.Glob(filepath.Join(config.ConfigDir, "*toml"))
	if err != nil {
		runErrors = append(runErrors, err)
		return runErrors
	}
	for _, p := range paths {
		log.Debug("Processing template resource " + p)
		t, err := New(p, config)
		if err != nil {
			runErrors = append(runErrors, err)
			log.Error(err.Error())
			continue
		}
		if err := t.process(); err != nil {
			runErrors = append(runErrors, err)
			log.Error(err.Error())
			continue
		}
		log.Debug("Processing of template resource " + p + " complete")
	}
	return runErrors
}
