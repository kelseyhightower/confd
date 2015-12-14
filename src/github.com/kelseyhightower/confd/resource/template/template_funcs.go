package template

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"time"
)

// loop take the number of elements to generate as well as the starting element. So loop(3, 5) will generate [5,6,7]
// It's helpful as it allow you to write templates like:
// {{range $index, $val := loop1to 2}}
// cpu-map {{ $val }} {{ $index }}
// {{end}}
// =>
// cpu-map 1 0
// cpu-map 2 1
func loop(n, s int) (arr []int) {
	arr = make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = i + s
	}
	return
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
	m["replace"] = strings.Replace
	m["add"] = func(a, b int) int { return a + b }
	m["sub"] = func(a, b int) int { return a - b }
	m["div"] = func(a, b int) int { return a / b }
	m["mod"] = func(a, b int) int { return a % b }
	m["mul"] = func(a, b int) int { return a * b }
	m["modBool"] = func(a, b int) bool { return a%b == 0 }
	m["loop"] = loop
	m["loop1to"] = func(n int) []int { return loop(n, 1) }
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
