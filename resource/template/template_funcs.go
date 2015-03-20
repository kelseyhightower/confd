package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func newFuncMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["base"] = path.Base
	m["split"] = strings.Split
	m["json"] = UnmarshalJsonObject
	m["jsonArray"] = UnmarshalJsonArray
	m["jsonGet"] = JsonGet
	m["dir"] = path.Dir
	m["getenv"] = os.Getenv
	m["join"] = strings.Join
	m["datetime"] = time.Now
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

func JsonGet(data string, property string) (interface{}, error) {
	var f interface{}
	err := json.Unmarshal([]byte(data), &f)
	if err != nil {
		panic(err)
	}

	m := f.(map[string]interface{})
	return getValue(m, strings.Split(property, "."))
}

func getValue(value interface{}, props []string) (val interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	if len(props) == 0 {
		return value, nil
	}
	first := props[0]
	rest := props[1:]
	switch vv := value.(type) {
	case map[string]interface{}:
		v := value.(map[string]interface{})[first]
		return getValue(v, rest)
	case []interface{}:
		values := value.([]interface{})
		if first == "*" {
			vs := make([]interface{}, len(values))
			for i, v := range values {
				vs[i], _ = getValue(v, rest)
			}
			return vs, nil
		}
		i, err := strconv.ParseInt(first, 10, 0)
		if err != nil {
			return nil, err
		}
		return getValue(values[i], rest)
	default:
		err := fmt.Errorf("Unsupported type: %v, for value: %#v", vv, value)
		return value, err
	}
}
