package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/confd/config"
	"github.com/kelseyhightower/confd/log"
)

// IsFileExist reports whether path exits.
func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Dump object
func Dump(v interface{}) {
	if v == nil {
		return
	}
	s := reflect.ValueOf(v).Elem()
	typeOfT := s.Type()

	log.Debug(typeOfT.String())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		log.Debug(fmt.Sprintf("%d: %s %s = '%v'", i, typeOfT.Field(i).Name, f.Type(), f.Interface()))
	}
}

func GetGlobalConfig(configFile string) (*config.GlobalConfig, error) {
	// default values
	gcf := &config.GlobalConfigFile{config.NewGlobalConfig()}

	// Update config from the TOML configuration file.
	log.Debug("Loading " + configFile)
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	_, err = toml.Decode(string(configBytes), gcf)
	if err != nil {
		return nil, err
	}

	return gcf.GlobalConfig, nil
}

func GetTemplateConfigs(configDir string) ([]*config.TemplateConfig, error) {
	log.Debug("Loading template configurations from confdir " + configDir)
	if !IsFileExist(configDir) {
		return nil, fmt.Errorf("Cannot load template configurations confdir '%s' does not exist", configDir)
	}

	// define paths
	tomlDir := filepath.Join(configDir, "conf.d")
	templatesDir := filepath.Join(configDir, "templates")

	paths, err := RecursiveFindFiles(tomlDir, "*toml")
	if err != nil {
		return nil, err
	}

	templates := make([]*config.TemplateConfig, 0)
	for _, path := range paths {
		var tcf *config.TemplateConfigFile
		log.Debug("Loading template resource from " + path)
		_, err := toml.DecodeFile(path, &tcf)
		if err != nil {
			return nil, fmt.Errorf("Cannot process template config %s - %s", path, err.Error())
		}

		// prepend templates dir to src
		tc := &tcf.TemplateConfig
		tc.Src = filepath.Join(templatesDir, tc.Src)

		// add to the output
		templates = append(templates, tc)
	}

	return templates, nil
}

func GetBackendConfig(backend string, configFile string) (config.BackendConfig, error) {
	var bcf config.BackendConfigFile = nil
    switch backend {
    case "consul":
	    bcf = &config.ConsulBackendConfigFile{config.NewConsulBackendConfig()}
    case "env":
	    bcf = &config.EnvBackendConfigFile{config.NewEnvBackendConfig()}
    case "etcd":
	    bcf = &config.EtcdBackendConfigFile{config.NewEtcdBackendConfig()}
    case "redis":
	    bcf = &config.RedisBackendConfigFile{config.NewRedisBackendConfig()}
    case "zookeeper":
	    bcf = &config.ZookeeperBackendConfigFile{config.NewZookeeperBackendConfig()}
    case "fs":
	    bcf = &config.FsBackendConfigFile{config.NewFsBackendConfig()}
	default:
	    panic("invalid backend, this should never happen!")
	}

	// Update config from the TOML configuration file.
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	_, err = toml.Decode(string(configBytes), bcf)
	if err != nil {
		return nil, err
	}

	return bcf.ConfigFile(), nil
}

// Get Backend Node From SRV
func GetBackendNodesFromSRV(backend, srv string) ([]string, error) {
	nodes := make([]string, 0)

	// Split, if domain was found
	srvScheme := "http"
	srvTarget := srv
	parts := strings.SplitN(srv, ",", 2)
	if len(parts) == 2 {
		srvScheme = parts[0]
		srvTarget = parts[1]
	}

	// Ignore the CNAME as we don't need it.
	_, addrs, err := net.LookupSRV(backend, "tcp", srvTarget)
	if err != nil {
		return nodes, err
	}

	// Generate node from SRV output
	for _, srv := range addrs {
		host := strings.TrimRight(srv.Target, ".")
		port := strconv.FormatUint(uint64(srv.Port), 10)
		nodes = append(nodes, fmt.Sprintf("%s://%s", srvScheme, net.JoinHostPort(host, port)))
	}

	return nodes, nil
}

// Default to values provided by functor
func GetBackendNodesFromSRVOrElse(backend, srv string, defNodes func() []string) []string {
	if nodes, err := GetBackendNodesFromSRV(backend, srv); err == nil && len(nodes) > 0 {
		return nodes
	}
	return defNodes()
}

// Find files with pattern in the root with depth.
func RecursiveFindFiles(root string, pattern string) ([]string, error) {
	files := make([]string, 0)
	findfile := func(path string, f os.FileInfo, err error) (inner error) {
		if err != nil {
			return
		}
		if f.IsDir() {
			return
		} else if match, innerr := filepath.Match(pattern, f.Name()); innerr == nil && match {
			files = append(files, path)
		}
		return
	}
	err := filepath.Walk(root, findfile)
	if len(files) == 0 {
		return files, err
	} else {
		return files, err
	}
}
