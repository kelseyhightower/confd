package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/etcd"
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

type templateConfig struct {
	Template Template
}

type FileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode uint32
	Md5  string
}

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

func ProcessTemplateConfigs() error {
	paths, err := filepath.Glob(filepath.Join(config.ConfigDir(), "*toml"))
	if err != nil {
		return err
	}
	for _, p := range paths {
		var tc *templateConfig
		_, err := toml.DecodeFile(p, &tc)
		if err != nil {
			return err
		}
		if err := tc.Template.Process(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Template) setVars() error {
	var err error
	t.Vars, err = etcd.GetValues(t.Keys)
	if err != nil {
		return err
	}
	return nil
}

func (t *Template) setStageFile() error {
	t.Src = filepath.Join(config.TemplateDir(), t.Src)
	if !IsFileExist(t.Src) {
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
	mode, _ := strconv.ParseUint(t.Mode, 0, 32)
	os.Chmod(temp.Name(), os.FileMode(mode))
	os.Chown(temp.Name(), t.Uid, t.Gid)
	t.StageFile = temp
	return nil
}

func (t *Template) sync() error {
	staged := t.StageFile.Name()
	defer os.Remove(staged)
	err, ok := SameFile(staged, t.Dest)
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
		if err := t.reload(); err != nil {
			return err
		}
	} else {
		log.Info(t.Dest + " in sync")
	}
	return nil
}

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

func (t *Template) reload() error {
	log.Debug("Running " + t.ReloadCmd)
	c := exec.Command("/bin/sh", "-c", t.ReloadCmd)
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Template) Process() error {
	if err := t.setVars(); err != nil {
		return err
	}
	if err := t.setStageFile(); err != nil {
		return err
	}
	if err := t.sync(); err != nil {
		return err
	}
	return nil
}

func FileStat(name string) (fi FileInfo, err error) {
	if IsFileExist(name) {
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

func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}

func SameFile(src, dest string) (error, bool) {
	if !IsFileExist(dest) {
		return nil, false
	}
	d, err := FileStat(dest)
	if err != nil {
		return err, false
	}
	s, err := FileStat(src)
	if err != nil {
		return err, false
	}
	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Md5 != s.Md5 {
		return nil, false
	}
	return nil, true
}
