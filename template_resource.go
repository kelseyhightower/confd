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

// TemplateResourceConfig holds the parsed template resource.
type TemplateResourceConfig struct {
	TemplateResource TemplateResource `toml:"template"`
}

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	Dest      string
	FileMode  os.FileMode
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

// setVars sets the Vars for template resource.
func (t *TemplateResource) setVars() error {
	c, err := newEtcdClient(EtcdNodes(), ClientCert(), ClientKey())
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
// StageFile for the template resource.
// It returns an error if any.
func (t *TemplateResource) createStageFile() error {
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
// It returns an error if any.
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
// It return an error if any.
func ProcessTemplateResources() error {
	paths, err := filepath.Glob(filepath.Join(ConfigDir(), "*toml"))
	if err != nil {
		return err
	}
	for _, p := range paths {
		var tc *TemplateResourceConfig
		_, err := toml.DecodeFile(p, &tc)
		if err != nil {
			return err
		}
		if err := tc.TemplateResource.process(); err != nil {
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
