package clconf_test

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/imdario/mergo"
	"gitlab.com/pastdev/s2i/clconf/clconf"
)

const configMap = "" +
	"---\n" +
	"key: foobar\n" +
	"database:\n" +
	"  host: 127.0.0.1\n" +
	"  port: \"3306\"\n" +
	"upstream:\n" +
	"  app1: 10.0.1.10:8080\n" +
	"  app2: 10.0.1.11:8080\n" +
	"prefix:\n" +
	"  database:\n" +
	"    host: 127.0.0.1\n" +
	"    port: \"3306\"\n" +
	"  upstream:\n" +
	"    app1: 10.0.1.10:8080\n" +
	"    app2: 10.0.1.11:8080\n"
const secrets = "" +
	"---\n" +
	"database:\n" +
	"  password: p@sSw0rd\n" +
	"  username: confd\n" +
	"prefix:\n" +
	"  database:\n" +
	"    password: p@sSw0rd\n" +
	"    username: confd\n"
const configMapAndSecrets = "" +
	"---\n" +
	"key: foobar\n" +
	"database:\n" +
	"  host: 127.0.0.1\n" +
	"  password: p@sSw0rd\n" +
	"  port: \"3306\"\n" +
	"  username: confd\n" +
	"upstream:\n" +
	"  app1: 10.0.1.10:8080\n" +
	"  app2: 10.0.1.11:8080\n" +
	"prefix:\n" +
	"  database:\n" +
	"    host: 127.0.0.1\n" +
	"    password: p@sSw0rd\n" +
	"    port: \"3306\"\n" +
	"    username: confd\n" +
	"  upstream:\n" +
	"    app1: 10.0.1.10:8080\n" +
	"    app2: 10.0.1.11:8080\n"
const yaml1 = "" +
	"a: Nope\n" +
	"b:\n" +
	"  c: 2\n"
const yaml2 = "" +
	"a: Yup\n" +
	"b:\n" +
	"  e: 2\n" +
	"  f:\n" +
	"    g: foobar\n"
const yaml1and2 = "" +
	"a: Yup\n" +
	"b:\n" +
	"  c: 2\n" +
	"  e: 2\n" +
	"  f:\n" +
	"    g: foobar\n"
const yaml2and1 = "" +
	"a: Nope\n" +
	"b:\n" +
	"  c: 2\n" +
	"  e: 2\n" +
	"  f:\n" +
	"    g: foobar\n"

func TestBase64Strings(t *testing.T) {
	expected := []string{}
	encoded := []string{}
	actual, err := clconf.DecodeBase64Strings(encoded...)
	if err != nil || len(actual) != 0 {
		t.Errorf("Base64Strings empty failed: [%v]", actual)
	}

	expected = []string{"one", "two"}
	encoded = []string{
		base64.StdEncoding.EncodeToString([]byte(expected[0])),
		base64.StdEncoding.EncodeToString([]byte(expected[1]))}
	actual, err = clconf.DecodeBase64Strings(encoded...)
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("Base64Strings one two failed: [%v] == [%v]", expected, actual)
	}

	if _, err := clconf.DecodeBase64Strings("&*INVALID*&"); err == nil {
		t.Error("Base64Strings invalid should have failed")
	}
}

func TestFillValue(t *testing.T) {
	conf, _ := clconf.UnmarshalYaml(yaml1and2)

	type Eff struct {
		G string
	}
	type Bee struct {
		C int
		E int
		F Eff
	}
	type Root struct {
		A string
		B Bee
	}
	var root Root
	ok := clconf.FillValue("", conf, &root)
	expectedRoot := Root{A: "Yup", B: Bee{C: 2, E: 2, F: Eff{G: "foobar"}}}
	if !ok || !reflect.DeepEqual(expectedRoot, root) {
		t.Errorf("FillValue empty path failed: [%v] [%v] == [%v]", ok, expectedRoot, root)
	}

	var bee Bee
	ok = clconf.FillValue("b", conf, &bee)
	expectedBee := Bee{C: 2, E: 2, F: Eff{G: "foobar"}}
	if !ok || !reflect.DeepEqual(expectedBee, bee) {
		t.Errorf("FillValue first level failed: [%v] [%v] == [%v]", ok, expectedBee, bee)
	}

	type BeeLight struct {
		C int
		E int
	}
	var beeLight BeeLight
	ok = clconf.FillValue("b", conf, &beeLight)
	expectedBeeLight := BeeLight{C: 2, E: 2}
	if !ok || !reflect.DeepEqual(expectedBeeLight, beeLight) {
		t.Errorf("FillValue first level string failed: [%v] [%v] == [%v]", ok, expectedBeeLight, beeLight)
	}

	type Zee struct {
		ShouldntWork string
	}
	var zee Zee
	ok = clconf.FillValue("/z", conf, &zee)
	if ok {
		t.Error("FillValue invalid path should have failed")
	}

	ok = clconf.FillValue("a/", conf, &zee)
	if ok {
		t.Error("FillValue a should not have been z")
	}
}

func TestGetValue(t *testing.T) {
	conf, _ := clconf.UnmarshalYaml(yaml1and2)

	value, ok := clconf.GetValue(conf, "")
	if !ok || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", ok, conf, value)
	}

	value, ok = clconf.GetValue(conf, "/")
	if !ok || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", ok, conf, value)
	}

	value, ok = clconf.GetValue(conf, "/a")
	if !ok || value != "Yup" {
		t.Errorf("GetValue first level string failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue(conf, "/b//f//g")
	if !ok || value != "foobar" {
		t.Errorf("GetValue third level string multi slash failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue(conf, "/a/f")
	if ok {
		t.Errorf("GetValue non map indexing should have failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue(conf, "/z")
	if ok {
		t.Errorf("GetValue missing have failed: [%v] [%v]", ok, value)
	}
}

func TestLoadConf(t *testing.T) {
	envVars := []string{"a"}
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
		for _, name := range envVars {
			os.Unsetenv(name)
		}
		os.Unsetenv("YAML_FILES")
		os.Unsetenv("YAML_VARS")
	}()

	envValues := []string{base64.StdEncoding.EncodeToString([]byte("a: env"))}
	for index, name := range envVars {
		os.Setenv(name, envValues[index])
	}

	fileVars := []string{path.Join(tempDir, "a")}
	fileValues := []string{"a: file"}
	for index, name := range fileVars {
		ioutil.WriteFile(name, []byte(fileValues[index]), 0700)
	}

	overrides := []string{base64.StdEncoding.EncodeToString([]byte("a: override"))}

	actual, err := clconf.LoadConf([]string{}, []string{})
	if err != nil || len(actual) > 0 {
		t.Errorf("LoadConf no config failed")
	}

	expected, _ := clconf.UnmarshalYaml("a: override")
	actual, err = clconf.LoadConf([]string{}, overrides)
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf overrides only failed: [%v] != [%v]", expected, actual)
	}

	os.Setenv("YAML_FILES", fileVars[0])
	expected, _ = clconf.UnmarshalYaml(fileValues[0])
	actual, err = clconf.LoadConfFromEnvironment([]string{}, []string{})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf files only failed: [%v] != [%v]", expected, actual)
	}
	os.Unsetenv("YAML_FILES")

	os.Setenv("YAML_VARS", envVars[0])
	base64Values, err := clconf.DecodeBase64Strings(envValues[0])
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf env only failed decode: [%v]", err)
	}
	expected, _ = clconf.UnmarshalYaml(base64Values...)
	actual, err = clconf.LoadConfFromEnvironment([]string{}, []string{})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf env only failed: [%v] != [%v]", expected, actual)
	}
	os.Unsetenv("YAML_VARS")

	os.Setenv("YAML_FILES", fileVars[0])
	os.Setenv("YAML_VARS", envVars[0])
	expected, _ = clconf.UnmarshalYaml("a: override")
	actual, err = clconf.LoadConf([]string{}, overrides)
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf all failed: [%v] != [%v]", expected, actual)
	}
}

func TestMarshalYaml(t *testing.T) {
	value := map[interface{}]interface{}{"a": "b"}
	yaml, err := clconf.MarshalYaml(value)
	if err != nil || string(yaml) != "a: b\n" {
		t.Errorf("Marshal failed for [%v]: [%v] [%v]", value, err, yaml)
	}
}

func TestReadEnvVars(t *testing.T) {
	actual := clconf.ReadEnvVars()
	if len(actual) > 0 {
		t.Errorf("ReadEnvVars empty failed")
	}
}

func TestReadEnvVarsDoesNotExist(t *testing.T) {
	defer func() {
		recover()
	}()
	clconf.ReadEnvVars("NOT_AN_ENV_VAR_OR_PROBABLY_SHOULDNT_BE")
	t.Errorf("ReadEnvVars does not exist should have paniced")
}

func TestReadEnvVarsTempValues(t *testing.T) {
	names := []string{"FOO", "BAZ"}
	values := []string{"bar", "qux"}
	defer func() {
		for _, name := range names {
			os.Unsetenv(name)
		}
	}()

	for index, name := range names {
		os.Setenv(name, values[index])
	}
	actual := clconf.ReadEnvVars(names...)
	if !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadEnvVars FOO BAZ failed: [%v] [%v]", values, actual)
	}
}

func TestReadFiles(t *testing.T) {
	actual, err := clconf.ReadFiles()
	if err != nil || len(actual) > 0 {
		t.Errorf("ReadFiles empty failed")
	}
	if _, err := clconf.ReadFiles("NOT_A_FILE_OR_PROBABLY_SHOULDNT_BE"); err == nil {
		t.Errorf("ReadFiles does not exist should have paniced")
	}
}

func TestReadFilesDoesNotExist(t *testing.T) {
	defer func() {
		recover()
	}()
}

func TestReadFilesTempValues(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	names := []string{path.Join(tempDir, "foo"), path.Join(tempDir, "baz")}
	values := []string{"bar", "qux"}
	for index, name := range names {
		ioutil.WriteFile(name, []byte(values[index]), 0700)
	}
	actual, err := clconf.ReadFiles(names...)
	if err != nil || !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadFiles foo baz failed: [%v] [%v]", values, actual)
	}
}

func TestSaveConf(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	config := map[interface{}]interface{}{"a": "b"}
	file := filepath.Join(tempDir, "config.yml")
	err = clconf.SaveConf(config, file)
	if err != nil {
		t.Errorf("SafeConf failed: %v", err)
	}
	actual, err := ioutil.ReadFile(file)
	if err != nil {
		t.Errorf("SafeConf failed, unable to read %v", file)
	}
	if "a: b\n" != string(actual) {
		t.Errorf("SafeConf failed, unexpected config: %v", string(actual))
	}
}

func TestSetValue(t *testing.T) {
	expected := map[interface{}]interface{}{}
	actual := map[interface{}]interface{}{}
	err := clconf.SetValue(actual, "", "baz")
	if err == nil {
		t.Error("SetValue empty config no path should have failed")
	}

	expected = map[interface{}]interface{}{
		"foo": map[interface{}]interface{}{"bar": "baz"}}
	actual = map[interface{}]interface{}{}
	err = clconf.SetValue(actual, "/foo/bar", "baz")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue empty config failed [%v] != [%v]: %v", expected, actual, err)
	}

	actual = map[interface{}]interface{}{"foo": "bar"}
	err = clconf.SetValue(actual, "/foo/bar", "baz")
	if err == nil {
		t.Error("SetValue non map parent should have failed")
	}

	expected = map[interface{}]interface{}{"foo": "baz"}
	actual = map[interface{}]interface{}{"foo": "bar"}
	err = clconf.SetValue(actual, "/foo", "baz")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue replace value [%v] != [%v]: %v", expected, actual, err)
	}

	expected = map[interface{}]interface{}{"foo": "bar", "hip": "hop"}
	actual = map[interface{}]interface{}{"foo": "bar"}
	err = clconf.SetValue(actual, "/hip", "hop")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue add value [%v] != [%v]: %v", expected, actual, err)
	}
}

func testToKvMap(t *testing.T, input, expected interface{}, message string) {
	actual := clconf.ToKvMap(input)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("ToKvMap %s failed: [%v] != [%v]", message, expected, actual)
	}
}

func TestToKvMap(t *testing.T) {
	testToKvMap(t, nil, map[string]string{"/": ""}, "nil")
	testToKvMap(t, "foo", map[string]string{"/": "foo"}, "string")
	testToKvMap(t, 2, map[string]string{"/": "2"}, "number")
	testToKvMap(t,
		map[interface{}]interface{}{
			"a": "b",
			"c": 2,
		},
		map[string]string{
			"/a": "b",
			"/c": "2",
		}, "simple map")
	testToKvMap(t,
		map[interface{}]interface{}{
			"a": "b",
			"c": 2,
			"d": map[interface{}]interface{}{
				"e": "f",
				"g": 2,
			},
		},
		map[string]string{
			"/a":   "b",
			"/c":   "2",
			"/d/e": "f",
			"/d/g": "2",
		}, "multi-level map")
	testToKvMap(t,
		map[interface{}]interface{}{
			"a": "b",
			"c": 2,
			"d": map[interface{}]interface{}{
				"e": []interface{}{"f", 2, 2.2},
			},
		},
		map[string]string{
			"/a":       "b",
			"/c":       "2",
			"/d/e/f":   "",
			"/d/e/2":   "",
			"/d/e/2.2": "",
		}, "multi-level map with array")
}

func TestUnmarshalYaml(t *testing.T) {
	//_, err := clconf.UnmarshalYaml("foo")
	//if err == nil {
	//	t.Error("Unmarshal illegal char")
	//}

	//expected, _ := clconf.UnmarshalYaml(yaml2and1)
	//merged, err := clconf.UnmarshalYaml(yaml2, yaml1)
	//if err != nil || !reflect.DeepEqual(merged, expected) {
	//	t.Errorf("Merge 2 and 1 failed: [%v] != [%v]", expected, merged)
	//}

	expected, _ := clconf.UnmarshalYaml(configMapAndSecrets)
	merged, err := clconf.UnmarshalYaml(configMap, secrets)
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("ConfigMap and Secrets failed: [%v] != [%v]", expected, merged)
	}
}

func TestMerge(t *testing.T) {
	result := make(map[interface{}]interface{})

	configMap := map[interface{}]interface{}{
		"foo": "bar",
		"database": map[interface{}]interface{}{
			"hostname": "localhost",
			"port":     3306,
			"username": "admin",
		},
	}
	secrets := map[interface{}]interface{}{
		"hip": "hop",
		"database": map[interface{}]interface{}{
			"password": "p@ssw0rD",
			"username": "notadmin",
		},
	}

	if err := mergo.Merge(&result, secrets); err != nil {
		t.Errorf("merge failed: [%v]", err)
	}
	if !reflect.DeepEqual(result, secrets) {
		t.Errorf("merge incorrect: [%v] != [%v]", result, secrets)
	}

	expected := map[interface{}]interface{}{
		"foo": "bar",
		"hip": "hop",
		"database": map[interface{}]interface{}{
			"hostname": "localhost",
			"password": "p@ssw0rD",
			"port":     3306,
			"username": "notadmin",
		},
	}

	if err := mergo.Merge(&result, configMap); err != nil {
		t.Errorf("merge failed: [%v]", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("merge incorrect: [%v] != [%v]", result, expected)
	}
}
