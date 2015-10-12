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

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/util"
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
	config        *config.TemplateConfig
	fileMode      os.FileMode
	stageFile     *os.File
	funcMap       map[string]interface{}
	lastIndex     uint64
	store         memkv.Store
	storeClient   backends.StoreClient
}

var ErrEmptySrc = errors.New("empty src template")

// NewTemplateResource creates a TemplateResource.
func NewTemplateResource(config *config.TemplateConfig, storeClient backends.StoreClient) *TemplateResource {
	store := memkv.New()
	funcMap := newFuncMap()
	addFuncs(funcMap, store.FuncMap)

	tr := &TemplateResource{
		config:      config,
		funcMap:     funcMap,
		lastIndex:   0,
		store:       store,
		storeClient: storeClient,
	}

	return tr
}

// setVars sets the Vars for template resource.
func (t *TemplateResource) setVars() error {
	var err error
	log.Debug("Retrieving keys from store")
	log.Debug("Key prefix set to " + t.config.Prefix)
	result, err := t.storeClient.GetValues(appendPrefix(t.config.Prefix, t.config.Keys))
	if err != nil {
		return err
	}
	t.store.Purge()
	for k, v := range result {
		t.store.Set(filepath.Join("/", strings.TrimPrefix(k, t.config.Prefix)), v)
	}
	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (t *TemplateResource) createStageFile() error {
	log.Debug("Using source template " + t.config.Src)

	if !util.IsFileExist(t.config.Src) {
		return errors.New("Missing template: " + t.config.Src)
	}

	log.Debug("Compiling source template " + t.config.Src)
    tmpl, err := template.New(path.Base(t.config.Src)).Funcs(t.funcMap).ParseFiles(t.config.Src)
    if err != nil {
		return fmt.Errorf("Unable to process template %s, %s", t.config.Src, err)
	}

    // create TempFile in Dest directory to avoid cross-filesystem issues
	temp, err := ioutil.TempFile(filepath.Dir(t.config.Dest), "."+filepath.Base(t.config.Dest))
	if err != nil {
		return err
	}

	if err = tmpl.Execute(temp, nil); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return err
	}
	defer temp.Close()
	
	// Set the owner, group, and mode on the stage file now to make it easier to
	// compare against the destination configuration file later.
	os.Chmod(temp.Name(), t.fileMode)
	os.Chown(temp.Name(), t.config.Uid, t.config.Gid)
	t.stageFile = temp
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *TemplateResource) sync(noop bool) error {
	staged := t.stageFile.Name()
	if t.config.KeepStageFile {
		log.Info("Keeping staged file: " + staged)
	} else {
		defer os.Remove(staged)
	}

	log.Debug("Comparing candidate config to " + t.config.Dest)
	ok, err := sameConfig(staged, t.config.Dest)
	if err != nil {
		log.Error(err.Error())
	}

	if noop {
		log.Warning("Noop mode enabled. " + t.config.Dest + " will not be modified")
		return nil
	}

	if !ok {
		log.Info("Target config " + t.config.Dest + " out of sync")
		if t.config.CheckCmd != "" {
			if err := t.check(); err != nil {
				return errors.New("Config check failed: " + err.Error())
			}
		}
		log.Debug("Overwriting target config " + t.config.Dest)
		err := os.Rename(staged, t.config.Dest)
		if err != nil {
			if strings.Contains(err.Error(), "device or resource busy") {
				log.Debug("Rename failed - target is likely a mount.config. Trying to write instead")
				// try to open the file and write to it
				var contents []byte
				var rerr error
				contents, rerr = ioutil.ReadFile(staged)
				if rerr != nil {
					return rerr
				}
				err := ioutil.WriteFile(t.config.Dest, contents, t.fileMode)
				// make sure owner and group match the temp file, in case the file was created with WriteFile
				os.Chown(t.config.Dest, t.config.Uid, t.config.Gid)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		if t.config.ReloadCmd != "" {
			if err := t.reload(); err != nil {
				return err
			}
		}
		log.Info("Target config " + t.config.Dest + " has been updated")
	} else {
		log.Debug("Target config " + t.config.Dest + " in sync")
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
	data["src"] = t.stageFile.Name()
	tmpl, err := template.New("checkcmd").Parse(t.config.CheckCmd)
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
		log.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	log.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (t *TemplateResource) reload() error {
	log.Debug("Running " + t.config.ReloadCmd)
	c := exec.Command("/bin/sh", "-c", t.config.ReloadCmd)
	output, err := c.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("%q", string(output)))
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
func (t *TemplateResource) process(noop bool) error {
	if err := t.setFileMode(); err != nil {
		return err
	}
	if err := t.setVars(); err != nil {
		return err
	}
	if err := t.createStageFile(); err != nil {
		return err
	}
	if err := t.sync(noop); err != nil {
		return err
	}
	return nil
}

// setFileMode sets the FileMode.
func (t *TemplateResource) setFileMode() error {
	if t.config.Mode == "" {
		if !util.IsFileExist(t.config.Dest) {
			t.fileMode = 0644
		} else {
			fi, err := os.Stat(t.config.Dest)
			if err != nil {
				return err
			}
			t.fileMode = fi.Mode()
		}
	} else {
		mode, err := strconv.ParseUint(t.config.Mode, 0, 32)
		if err != nil {
			return err
		}
		t.fileMode = os.FileMode(mode)
	}
	return nil
}
