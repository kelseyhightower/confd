package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template/extensions"
)

// RegisterPlugin register a new function to be used inside a template
func RegisterPlugin(name string, plugin interface{}) []string {
	return extensions.RegisterExtension(plugin, name)
}

func newFuncMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["base"] = path.Base
	m["split"] = strings.Split
	m["json"] = UnmarshalJsonObject
	m["jsonArray"] = UnmarshalJsonArray
	m["dir"] = path.Dir
	m["getenv"] = os.Getenv
	m["join"] = strings.Join
	m["datetime"] = time.Now
	m["toUpper"] = strings.ToUpper
	m["toLower"] = strings.ToLower
	m["contains"] = strings.Contains

	// register additional template functions (from plugins)
	for name, plugin := range extensions.TemplatePlugins.All() {
		log.Debug(fmt.Sprintf("adding plugin %v -> %v", name, plugin.Function))
		m[name] = plugin.Function()
	}

	return m
}

func addFuncs(out, in map[string]interface{}) {
	for name, fn := range in {
		out[name] = fn
	}
}

func UnmarshalJsonObject(data string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func UnmarshalJsonArray(data string) ([]interface{}, error) {
	var ret []interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}
