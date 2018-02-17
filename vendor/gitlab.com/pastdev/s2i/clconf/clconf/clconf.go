// Package clconf provides functions to extract values from a set of yaml
// files after merging them.
package clconf

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Splitter is the regex used to split YAML_FILES and YAML_VARS
var Splitter = regexp.MustCompile(`,`)

// DecodeBase64Strings will decode all the base64 strings supplied
func DecodeBase64Strings(values ...string) ([]string, error) {
	var contents []string
	for _, value := range values {
		content, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		contents = append(contents, string(content))
	}
	return contents, nil
}

// FillValue will fill a struct, out, with values from conf.
func FillValue(keyPath string, conf interface{}, out interface{}) bool {
	value, ok := GetValue(conf, keyPath)
	if !ok {
		return false
	}
	err := mapstructure.Decode(value, out)
	if err != nil {
		return false
	}
	return ok
}

// GetValue returns the value at the indicated path.  Paths are separated by
// the '/' character.  The empty string or "/" will return conf itself.
func GetValue(conf interface{}, keyPath string) (interface{}, bool) {
	if keyPath == "" {
		return conf, true
	}

	var value = conf
	for _, part := range strings.Split(keyPath, "/") {
		if part == "" {
			continue
		}
		if reflect.ValueOf(value).Kind() != reflect.Map {
			log.Warnf("value at [%v] not a map: %v", part, reflect.ValueOf(value).Kind())
			return nil, false
		}
		partValue, ok := value.(map[interface{}]interface{})[part]
		if !ok {
			log.Warnf("value at [%v] does not exist", part)
			return nil, false
		}
		value = partValue
	}
	return value, true
}

// LoadConf will load all configurations provided.  In order of precedence
// (highest last), files, overrides.
func LoadConf(files []string, overrides []string) (map[interface{}]interface{}, error) {
	yamls := []string{}
	if len(files) > 0 {
		moreYamls, err := ReadFiles(files...)
		if err != nil {
			return nil, err
		}
		yamls = append(yamls, moreYamls...)
	}
	if len(overrides) > 0 {
		moreYamls, err := DecodeBase64Strings(overrides...)
		if err != nil {
			return nil, err
		}
		yamls = append(yamls, moreYamls...)
	}

	return UnmarshalYaml(yamls...)
}

// LoadConfFromEnvironment will load all configurations present.  In order
// of precedence (highest last), files, YAML_FILES env var, overrides,
// YAML_VARS env var.
func LoadConfFromEnvironment(files []string, overrides []string) (map[interface{}]interface{}, error) {
	if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok {
		files = append(files, Splitter.Split(yamlFiles, -1)...)
	}
	if yamlVars, ok := os.LookupEnv("YAML_VARS"); ok {
		overrides = append(overrides, ReadEnvVars(Splitter.Split(yamlVars, -1)...)...)
	}
	return LoadConf(files, overrides)
}

// LoadSettableConfFromEnvironment loads configuration for setting.  Only one
// file is allowed, but can be specified, either by the environment variable
// YAML_FILES, or as the single value in the supplied files array.  Returns
// the name of the file to be written, the conf map, and a non-nil error upon
// failure.  If the file does not currently exist, an empty map will be returned
// and a call to SaveConf will create the file.
func LoadSettableConfFromEnvironment(files []string) (string, map[interface{}]interface{}, error) {
	if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok {
		files = append(files, Splitter.Split(yamlFiles, -1)...)
	}
	if len(files) != 1 {
		return "", nil, errors.New("Exactly one file required with setv")
	}

	if _, err := os.Stat(files[0]); os.IsNotExist(err) {
		return files[0], map[interface{}]interface{}{}, nil
	}

	config, err := LoadConf(files, []string{})
	return files[0], config, err
}

// MarshalYaml will convert an object to yaml
func MarshalYaml(in interface{}) ([]byte, error) {
	value, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// ReadEnvVars will read all the environment variables named and return an
// array of their values.  The order of the names to values will be
// preserved.
func ReadEnvVars(names ...string) []string {
	var values []string
	for _, name := range names {
		if value, ok := os.LookupEnv(name); ok {
			values = append(values, value)
		} else {
			log.Panicf("Read env var [%s] failed, does not exist", name)
		}
	}
	return values
}

// ReadFiles will read all the files supplied and return an array of their
// contents.  The order of files to contents will be preserved.
func ReadFiles(files ...string) ([]string, error) {
	var contents []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, err
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		contents = append(contents, string(content))
	}
	return contents, nil
}

func splitKeyPath(keyPath string) ([]string, string) {
	parts := []string{}

	for _, parentPart := range strings.Split(keyPath, "/") {
		if parentPart == "" {
			continue
		}
		parts = append(parts, parentPart)
	}

	lastIndex := len(parts) - 1
	if lastIndex >= 0 {
		return parts[:lastIndex], parts[lastIndex]
	}
	return parts, keyPath
}

// SetValue will set the value of config at keyPath to value
func SetValue(config interface{}, keyPath string, value interface{}) error {
	configMap, ok := config.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("Config not a map")
	}
	parentParts, key := splitKeyPath(keyPath)
	if key == "" {
		return fmt.Errorf("[%v] is an invalid keyPath", keyPath)
	}

	parent := configMap
	for _, parentPart := range parentParts {
		parentValue, ok := parent[parentPart]
		if !ok {
			parentValue = make(map[interface{}]interface{})
			parent[parentPart] = parentValue
		}
		valueMap, ok := parentValue.(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("Parent not a map")
		}

		parent = valueMap
	}

	parent[key] = value

	return nil
}

// SaveConf will save config to file as yaml
func SaveConf(config interface{}, file string) error {
	yamlBytes, err := MarshalYaml(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, yamlBytes, 0660)
}

// ToKvMap will return a one-level map of key value pairs where the key is
// a / separated path of subkeys.
func ToKvMap(conf interface{}) map[string]string {
	kvMap := make(map[string]string)
	Walk(func(keyStack []string, value interface{}) {
		key := "/" + strings.Join(keyStack, "/")
		if value == nil {
			kvMap[key] = ""
		} else {
			kvMap[key] = fmt.Sprintf("%v", value)
		}
	}, conf)
	return kvMap
}

func unmarshalYaml(yamlBytes ...[]byte) (map[interface{}]interface{}, error) {
	result := make(map[interface{}]interface{})
	for index := len(yamlBytes) - 1; index >= 0; index-- {
		yamlMap := make(map[interface{}]interface{})

		err := yaml.Unmarshal(yamlBytes[index], &yamlMap)
		if err != nil {
			return nil, err
		}

		if err := mergo.Merge(&result, yamlMap); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// UnmarshalYaml will parse all the supplied yaml strings, merge the resulting
// objects, and return the resulting map
func UnmarshalYaml(yamlStrings ...string) (map[interface{}]interface{}, error) {
	yamlBytes := make([][]byte, len(yamlStrings))
	for _, yaml := range yamlStrings {
		yamlBytes = append(yamlBytes, []byte(yaml))
	}
	return unmarshalYaml(yamlBytes...)
}

// Walk will recursively iterate over all the nodes of conf calling callback
// for each node.
func Walk(callback func(key []string, value interface{}), conf interface{}) {
	node, ok := conf.(map[interface{}]interface{})
	if !ok {
		callback([]string{}, conf)
	}
	walk(callback, node, []string{})
}

func walk(callback func(key []string, value interface{}), node map[interface{}]interface{}, keyStack []string) {
	for k, v := range node {
		keyStack := append(keyStack, k.(string))

		switch v.(type) {
		case map[interface{}]interface{}:
			walk(callback, v.(map[interface{}]interface{}), keyStack)
		case []interface{}:
			for _, j := range v.([]interface{}) {
				switch j.(type) {
				case map[interface{}]interface{}:
					walk(callback, j.(map[interface{}]interface{}), keyStack)
				default:
					callback(append(keyStack, fmt.Sprintf("%v", j)), "")
				}
			}
		default:
			callback(keyStack, v)
		}
	}
}
