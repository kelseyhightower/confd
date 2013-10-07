package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"text/template"
)

// templateConfig holds the parsed template resource.
type templateConfig struct {
	Template Template
}

// A fileInfo describes a configuration file and is returned by fileStat.
type fileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode uint32
	Md5  string
}

// Template is the representation of a parsed confd template resource.
type Template struct {
	Dest      string
	Gid       int
	Keys      []string
	Mode      string
	Uid       int
	ReloadCmd string `toml:"reload_cmd"`
	CheckCmd  string `toml:"check_cmd"`
	StageFile *os.File
	Src       string
	Vars      map[string]interface{}
}

// setVars sets the Vars for template config.
func (t *Template) setVars() error {
	c, err := newEtcdClient(EtcdNodes())
	if err != nil {
		return err
	}
	t.Vars, err = getValues(c, Prefix(), t.Keys)
	if err != nil {
		return err
	}
	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template config.
// It returns an error if any.
func (t *Template) createStageFile() error {
	t.Src = filepath.Join(TemplateDir(), t.Src)
	if !isFileExist(t.Src) {
		return errors.New("Missing template: " + t.Src)
	}
	temp, err := ioutil.TempFile("", "")
	if err != nil {
		os.Remove(temp.Name())
		return err
	}
	tmpl := template.Must(template.ParseFiles(t.Src))
	if err = tmpl.Execute(temp, t.Vars); err != nil {
		return err
	}
	// Set the owner, group, and mode on the stage file now to make it easier to
	// compare against the destination configuration file later.
	mode, _ := strconv.ParseUint(t.Mode, 0, 32)
	os.Chmod(temp.Name(), os.FileMode(mode))
	os.Chown(temp.Name(), t.Uid, t.Gid)
	t.StageFile = temp
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *Template) sync() error {
	staged := t.StageFile.Name()
	defer os.Remove(staged)
	err, ok := sameConfig(staged, t.Dest)
	if err != nil {
		log.Error(err.Error())
	}
	if !ok {
		log.Info(t.Dest + " not in sync")
		if t.CheckCmd != "" {
			if err := t.check(); err != nil {
				return errors.New("Config check failed: " + err.Error())
			}
		}
		os.Rename(staged, t.Dest)
		if t.ReloadCmd != "" {
			if err := t.reload(); err != nil {
				return err
			}
		}
	} else {
		log.Info(t.Dest + " in sync")
	}
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (t *Template) check() error {
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
func (t *Template) reload() error {
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
func (t *Template) process() error {
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

// ProcessTemplateConfigs is a convenience function that loads all the template
// config files and processes them serially. Called from the main function.
// It return an error if any.
func ProcessTemplateConfigs() error {
	paths, err := filepath.Glob(filepath.Join(ConfigDir(), "*toml"))
	if err != nil {
		return err
	}
	for _, p := range paths {
		var tc *templateConfig
		_, err := toml.DecodeFile(p, &tc)
		if err != nil {
			return err
		}
		if err := tc.Template.process(); err != nil {
			return err
		}
	}
	return nil
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
func sameConfig(src, dest string) (error, bool) {
	if !isFileExist(dest) {
		return nil, false
	}
	d, err := fileStat(dest)
	if err != nil {
		return err, false
	}
	s, err := fileStat(src)
	if err != nil {
		return err, false
	}
	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Md5 != s.Md5 {
		return nil, false
	}
	return nil, true
}
