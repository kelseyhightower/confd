// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package template

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/etcd/etcdutil"
	"github.com/kelseyhightower/confd/log"
)

// TemplateResourceConfig holds the parsed template resource.
type TemplateResourceConfig struct {
	TemplateResource TemplateResource `toml:"template"`
}

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	Dest       string
	FileMode   os.FileMode
	Gid        int
	Keys       []string
	Mode       string
	Uid        int
	ReloadCmd  string `toml:"reload_cmd"`
	CheckCmd   string `toml:"check_cmd"`
	StageFile  *os.File
	Src        string
	Vars       map[string]interface{}
	etcdClient etcdutil.EtcdClient
}

// NewTemplateResourceFromPath creates a TemplateResource using a decoded file path
// and the supplied EtcdClient as input.
// It returns a TemplateResource and an error if any.
func NewTemplateResourceFromPath(path string, c etcdutil.EtcdClient) (*TemplateResource, error) {
	if c == nil {
		return nil, errors.New("A valid EtcdClient is required.")
	}
	var tc *TemplateResourceConfig
	log.Debug("Loading template resource from " + path)
	_, err := toml.DecodeFile(path, &tc)
	if err != nil {
		return nil, fmt.Errorf("Cannot process template resource %s - %s", path, err.Error())
	}
	tc.TemplateResource.etcdClient = c
	return &tc.TemplateResource, nil
}

// setVars sets the Vars for template resource.
func (t *TemplateResource) setVars() error {
	var err error
	log.Debug("Retrieving keys from etcd")
	log.Debug("Key prefix set to " + config.Prefix())
	t.Vars, err = etcdutil.GetValues(t.etcdClient, config.Prefix(), t.Keys)
	if err != nil {
		return err
	}
	return nil
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
	temp, err := ioutil.TempFile("", "")
	if err != nil {
		os.Remove(temp.Name())
		return err
	}
	log.Debug("Compiling source template " + t.Src)
	tplFuncMap := make(template.FuncMap)
	tplFuncMap["Base"] = path.Base
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
	log.Debug("Running " + t.ReloadCmd)
	c := exec.Command("/bin/sh", "-c", t.ReloadCmd)
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

// process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from etcd, then we stage a candidate configuration file, and finally sync
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
func ProcessTemplateResources(c etcdutil.EtcdClient) []error {
	runErrors := make([]error, 0)
	var err error
	if c == nil {
		runErrors = append(runErrors, errors.New("An etcd client is required"))
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
		t, err := NewTemplateResourceFromPath(p, c)
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

// fileStat return a fileInfo describing the named file.
func fileStat(name string) (fi fileInfo, err error) {
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

// isFileExist reports whether path exits.
func isFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
