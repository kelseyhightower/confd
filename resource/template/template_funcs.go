package template

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"regexp"

	yaml "gopkg.in/yaml.v2"

	"github.com/kelseyhightower/memkv"
)

const (
	//FloatNumberPrecision float number precision
	FloatNumberPrecision = 0.0000001
)

var (
	timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
	kvType = reflect.TypeOf(memkv.KVPair{}).Kind()
)

func newFuncMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["base"] = path.Base
	m["split"] = strings.Split
	m["json"] = UnmarshalJsonObject
	m["jsonArray"] = UnmarshalJsonArray
	m["dir"] = path.Dir
	m["map"] = CreateMap
	m["getenv"] = Getenv
	m["join"] = strings.Join
	m["datetime"] = time.Now
	m["toUpper"] = strings.ToUpper
	m["toLower"] = strings.ToLower
	m["contains"] = strings.Contains
	m["replace"] = strings.Replace
	m["trimSuffix"] = strings.TrimSuffix
	m["lookupIP"] = LookupIP
	m["lookupV4IP"] = LookupV4IP
	m["lookupSRV"] = LookupSRV
	m["fileExists"] = isFileExist
	m["base64Encode"] = Base64Encode
	m["base64Decode"] = Base64Decode
	m["reverse"] = Reverse
	m["sortByLength"] = SortByLength
	m["sortKVByLength"] = SortKVByLength
	m["add"] = func(a, b interface{}) (interface{}, error) { return DoArithmetic(a, b, '+') }
	m["div"] = func(a, b interface{}) (interface{}, error) { return DoArithmetic(a, b, '/') }
	m["mul"] = func(a, b interface{}) (interface{}, error) { return DoArithmetic(a, b, '*') }
	m["sub"] = func(a, b interface{}) (interface{}, error) { return DoArithmetic(a, b, '-') }
	m["mod"] = mod
	m["max"] = max
	m["min"] = min
	m["eq"] = eq
	m["ne"] = ne
	m["gt"] = gt
	m["ge"] = ge
	m["lt"] = lt
	m["le"] = le
	m["seq"] = Seq
	m["filter"] = Filter
	m["toJson"] = ToJson
	m["toYaml"] = ToYaml
	return m
}

func addFuncs(out, in map[string]interface{}) {
	for name, fn := range in {
		out[name] = fn
	}
}

// Seq creates a sequence of integers. It's named and used as GNU's seq.
// Seq takes the first and the last element as arguments. So Seq(3, 5) will generate [3,4,5]
func Seq(first, last int) []int {
	var arr []int
	for i := first; i <= last; i++ {
		arr = append(arr, i)
	}
	return arr
}

type byLengthKV []memkv.KVPair

func (s byLengthKV) Len() int {
	return len(s)
}

func (s byLengthKV) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byLengthKV) Less(i, j int) bool {
	return len(s[i].Key) < len(s[j].Key)
}

func SortKVByLength(values []memkv.KVPair) []memkv.KVPair {
	sort.Sort(byLengthKV(values))
	return values
}

type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

func SortByLength(values []string) []string {
	sort.Sort(byLength(values))
	return values
}

//Reverse returns the array in reversed order
//works with []string and []KVPair
func Reverse(values interface{}) interface{} {
	switch values.(type) {
	case []string:
		v := values.([]string)
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	case []memkv.KVPair:
		v := values.([]memkv.KVPair)
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	}
	return values
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will the default value if the variable is not present.
// If no default value was given - returns "".
func Getenv(key string, v ...string) string {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}

	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// CreateMap creates a key-value map of string -> interface{}
// The i'th is the key and the i+1 is the value
func CreateMap(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid map call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("map keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
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

func lookupIP(data string, ipv4Only bool) []string {
	ips, err := net.LookupIP(data)
	if err != nil {
		return nil
	}
	// "Cast" IPs into strings and sort the array
	ipStrings := make([]string, 0, len(ips))

	for _, ip := range ips {
		if ipv4Only && ip.To4() == nil {
			continue
		}
		ipStrings = append(ipStrings, ip.String())
	}
	sort.Strings(ipStrings)
	return ipStrings
}

func LookupIP(data string) []string {
	return lookupIP(data, false)
}

func LookupV4IP(data string) []string {
	return lookupIP(data, true)
}

type sortSRV []*net.SRV

func (s sortSRV) Len() int {
	return len(s)
}

func (s sortSRV) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortSRV) Less(i, j int) bool {
	str1 := fmt.Sprintf("%s%d%d%d", s[i].Target, s[i].Port, s[i].Priority, s[i].Weight)
	str2 := fmt.Sprintf("%s%d%d%d", s[j].Target, s[j].Port, s[j].Priority, s[j].Weight)
	return str1 < str2
}

func LookupSRV(service, proto, name string) []*net.SRV {
	_, addrs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return []*net.SRV{}
	}
	sort.Sort(sortSRV(addrs))
	return addrs
}

func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func Base64Decode(data string) (string, error) {
	s, err := base64.StdEncoding.DecodeString(data)
	return string(s), err
}

// eq returns the boolean truth of arg1 == arg2.
func eq(x, y interface{}) bool {
	normalize := func(v interface{}) interface{} {
		vv := reflect.ValueOf(v)
		nv, err := stringToNumber(vv)
		if err == nil {
			vv = nv
		}
		switch vv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(vv.Int()) //may overflow
		case reflect.Float32, reflect.Float64:
			return vv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(vv.Uint()) //may overflow
		default:
			return v
		}
	}
	x = normalize(x)
	y = normalize(y)
	return reflect.DeepEqual(x, y)
}

// ne returns the boolean truth of arg1 != arg2.
func ne(x, y interface{}) bool {
	return !eq(x, y)
}

// ge returns the boolean truth of arg1 >= arg2.
func ge(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left >= right
}

// gt returns the boolean truth of arg1 > arg2.
func gt(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left > right
}

// le returns the boolean truth of arg1 <= arg2.
func le(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left <= right
}

// lt returns the boolean truth of arg1 < arg2.
func lt(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left < right
}

// mod returns a % b.
func mod(a, b interface{}) (int64, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	var err error
	if av.Kind() == reflect.String {
		av, err = stringToNumber(av)
		if err != nil {
			return 0, err
		}
	}
	if bv.Kind() == reflect.String {
		bv, err = stringToNumber(bv)
		if err != nil {
			return 0, err
		}
	}

	var ai, bi int64

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ai = av.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ai = int64(av.Uint()) //may overflow
	default:
		return 0, errors.New("Modulo operator can't be used with non integer value")
	}

	switch bv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bi = bv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bi = int64(bv.Uint()) //may overflow
	default:
		return 0, errors.New("Modulo operator can't be used with non integer value")
	}

	if bi == 0 {
		return 0, errors.New("The number can't be divided by zero at modulo operation")
	}

	return ai % bi, nil
}

// max returns the larger of a or b
func max(a, b interface{}) (float64, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	var err error
	if av.Kind() == reflect.String {
		av, err = stringToNumber(av)
		if err != nil {
			return 0, err
		}
	}
	if bv.Kind() == reflect.String {
		bv, err = stringToNumber(bv)
		if err != nil {
			return 0, err
		}
	}

	var af, bf float64

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		af = float64(av.Int()) //may overflow
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		af = float64(av.Uint()) //may overflow
	case reflect.Float64, reflect.Float32:
		af = av.Float()
	default:
		return 0, errors.New("Max operator can't apply to the values")
	}

	switch bv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bf = float64(bv.Int()) //may overflow
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bf = float64(bv.Uint()) //may overflow
	case reflect.Float64, reflect.Float32:
		bf = bv.Float()
	default:
		return 0, errors.New("Max operator can't apply to the values")
	}

	return math.Max(af, bf), nil
}

// min returns the smaller of a or b
func min(a, b interface{}) (float64, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	var err error
	if av.Kind() == reflect.String {
		av, err = stringToNumber(av)
		if err != nil {
			return 0, err
		}
	}
	if bv.Kind() == reflect.String {
		bv, err = stringToNumber(bv)
		if err != nil {
			return 0, err
		}
	}

	var af, bf float64

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		af = float64(av.Int()) //may overflow
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		af = float64(av.Uint()) //may overflow
	case reflect.Float64, reflect.Float32:
		af = av.Float()
	default:
		return 0, errors.New("Max operator can't apply to the values")
	}

	switch bv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bf = float64(bv.Int()) //may overflow
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bf = float64(bv.Uint()) //may overflow
	case reflect.Float64, reflect.Float32:
		bf = bv.Float()
	default:
		return 0, errors.New("Max operator can't apply to the values")
	}

	return math.Min(af, bf), nil
}

func compareGetFloat(a interface{}, b interface{}) (float64, float64) {
	var left, right float64
	var leftStr, rightStr *string
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		left = float64(av.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		left = float64(av.Int())
	case reflect.Float32, reflect.Float64:
		left = av.Float()
	case reflect.String:
		var err error
		left, err = strconv.ParseFloat(av.String(), 64)
		if err != nil {
			str := av.String()
			leftStr = &str
		}
	case reflect.Struct:
		switch av.Type() {
		case timeType:
			left = float64(toTimeUnix(av))
		}
	}

	bv := reflect.ValueOf(b)

	switch bv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		right = float64(bv.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		right = float64(bv.Int())
	case reflect.Float32, reflect.Float64:
		right = bv.Float()
	case reflect.String:
		var err error
		right, err = strconv.ParseFloat(bv.String(), 64)
		if err != nil {
			str := bv.String()
			rightStr = &str
		}
	case reflect.Struct:
		switch bv.Type() {
		case timeType:
			right = float64(toTimeUnix(bv))
		}
	}

	switch {
	case leftStr == nil || rightStr == nil:
	case *leftStr < *rightStr:
		return 0, 1
	case *leftStr > *rightStr:
		return 1, 0
	default:
		return 0, 0
	}

	return left, right
}

func toTimeUnix(v reflect.Value) int64 {
	if v.Kind() == reflect.Interface {
		return toTimeUnix(v.Elem())
	}
	if v.Type() != timeType {
		panic("coding error: argument must be time.Time type reflect Value")
	}
	return v.MethodByName("Unix").Call([]reflect.Value{})[0].Int()
}

// DoArithmetic performs arithmetic operations (+,-,*,/) using reflection to
// determine the type of the two terms.
// This func will auto convert string and uint to int64/float64, then apply  operations,
// return float64, or int64
func DoArithmetic(a, b interface{}, op rune) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	var ai, bi int64
	var af, bf float64

	var err error
	if av.Kind() == reflect.String {
		av, err = stringToNumber(av)
		if err != nil {
			return nil, err
		}
	}
	if bv.Kind() == reflect.String {
		bv, err = stringToNumber(bv)
		if err != nil {
			return nil, err
		}
	}

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ai = av.Int()
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bi = bv.Int()
		case reflect.Float32, reflect.Float64:
			af = float64(ai) // may overflow
			ai = 0
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bi = int64(bv.Uint()) // may overflow
		default:
			return nil, errors.New("Can't apply the operator to the values")
		}
	case reflect.Float32, reflect.Float64:
		af = av.Float()
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bf = float64(bv.Int()) // may overflow
		case reflect.Float32, reflect.Float64:
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bf = float64(bv.Uint()) // may overflow
		default:
			return nil, errors.New("Can't apply the operator to the values")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ai = int64(av.Uint()) // may overflow
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bi = bv.Int()
		case reflect.Float32, reflect.Float64:
			af = float64(ai) // may overflow
			ai = 0
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bi = int64(bv.Uint()) // may overflow
		default:
			return nil, errors.New("Can't apply the operator to the values")
		}
	default:
		return nil, errors.New("Can't apply the operator to the values")
	}

	switch op {
	case '+':
		if !isFloatZero(af) || !isFloatZero(bf) {
			return af + bf, nil
		} else if ai != 0 || bi != 0 {
			return ai + bi, nil
		}
		return 0, nil
	case '-':
		if !isFloatZero(af) || !isFloatZero(bf) {
			return af - bf, nil
		} else if ai != 0 || bi != 0 {
			return ai - bi, nil
		}
		return 0, nil
	case '*':
		if !isFloatZero(af) || !isFloatZero(bf) {
			return af * bf, nil
		}
		if ai != 0 || bi != 0 {
			return ai * bi, nil
		}
		return 0, nil
	case '/':
		if !isFloatZero(bf) {
			return af / bf, nil
		} else if bi != 0 {
			return ai / bi, nil
		}
		return nil, errors.New("Can't divide the value by 0")
	default:
		return nil, errors.New("There is no such an operation")
	}
}

func stringToNumber(value reflect.Value) (reflect.Value, error) {
	var result reflect.Value
	str := value.String()
	if isFloat(str) {
		vf, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Can't apply the operator to the value [%s] ,err [%s] ", str, err.Error())
		}
		result = reflect.ValueOf(vf)
	} else {
		vi, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Can't apply the operator to the value [%s] ,err [%s] ", str, err.Error())
		}
		result = reflect.ValueOf(vi)
	}
	return result, nil
}

func isFloat(value string) bool {
	return strings.Index(value, ".") >= 0
}

func isFloatZero(value float64) bool {
	return math.Abs(value)-0 < FloatNumberPrecision
}

func Filter(regex string, c interface{}) ([]interface{}, error) {
	cv := reflect.ValueOf(c)

	switch cv.Kind() {
	case reflect.Array, reflect.Slice:
		result := make([]interface{}, 0, cv.Len())
		for i := 0; i < cv.Len(); i++ {
			v := cv.Index(i)
			if v.Kind() == reflect.Interface {
				v = reflect.ValueOf(v.Interface())
			}
			if v.Kind() == reflect.String {
				matched, err := regexp.MatchString(regex, v.String())
				if err != nil {
					return nil, err
				}
				if matched {
					result = append(result, v.String())
				}
			} else if v.Kind() == kvType {
				kv := v.Interface().(memkv.KVPair)
				matched, err := regexp.MatchString(regex, kv.Value)
				if err != nil {
					return nil, err
				}
				if matched {
					result = append(result, kv)
				}
			}
		}
		return result, nil
	default:
		return nil, errors.New("filter only support slice or array.")
	}
}

func ToJson(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ToYaml(v interface{}) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
