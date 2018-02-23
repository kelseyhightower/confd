package clconf

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/urfave/cli"
)

func getCsetvOutcome(file string, args, options []string) (string, interface{}, interface{}, error) {
	_, expected, err := LoadSettableConfFromEnvironment([]string{file})
	if err != nil {
		return "", nil, nil, err
	}

	context := NewTestSetvContext(file, args, options)
	if err := csetv(context); err != nil {
		return "", nil, nil, err
	}
	path := getPath(context)
	value := context.Args().Get(1)

	SetValue(expected, path, value)
	actual, err := LoadConf([]string{file}, []string{})

	if err != nil {
		return path, nil, nil, err
	}
	secretAgent, err := newSecretAgentFromCli(context)
	if err != nil {
		return path, nil, nil, fmt.Errorf("newSecretAgentFromCli failed: %v", err)
	}
	if err := secretAgent.DecryptPaths(actual, path); err != nil {
		return path, nil, nil, fmt.Errorf("DecryptPaths failed: %v", err)
	}

	return path, expected, actual, nil
}

func getGetvOutcome(config interface{}, args []string, options []string) (*cli.Context, string, interface{}, interface{}, error) {
	var err error
	context := NewTestGetvContext(args, options)
	path := getPath(context)
	expected, ok := GetValue(config, path)
	if !ok {
		if defaultValue, defaultOk := getDefault(context); defaultOk {
			expected = defaultValue
			ok = true
		}
	}
	if decryptPaths := context.StringSlice("decrypt"); len(decryptPaths) > 0 {
		secretAgent, err := newSecretAgentFromCli(context)
		if err != nil {
			return context, path, nil, nil, fmt.Errorf("newSecretAgentFromCli failed: %v", err)
		}
		if stringValue, ok := expected.(string); ok {
			if len(decryptPaths) != 1 || !(decryptPaths[0] == "" || decryptPaths[0] == "/") {
				return context, path, nil, nil, errors.New("string value with non-root decrypt path")
			}
			decrypted, err := secretAgent.Decrypt(stringValue)
			if err != nil {
				return context, path, nil, nil, err
			}
			expected = decrypted
		} else {
			if err := secretAgent.DecryptPaths(expected, decryptPaths...); err != nil {
				return context, path, nil, nil, fmt.Errorf("DecryptPaths failed: %v", err)
			}
		}
	}

	_, actual, err := getValue(context)
	if ok && err != nil {
		return context, path, nil, nil, fmt.Errorf("getValue %s failed and shouldn't have: %v", path, err)
	} else if !ok && err == nil {
		return context, path, nil, nil, fmt.Errorf("getValue %s didn't fail and should have", path)
	}

	return context, path, expected, actual, nil
}

func getSetvOutcome(file string, args, options []string) (string, interface{}, interface{}, error) {
	_, expected, err := LoadSettableConfFromEnvironment([]string{file})
	if err != nil {
		return "", nil, nil, err
	}

	context := NewTestSetvContext(file, args, options)
	if err := setv(context); err != nil {
		return "", nil, nil, err
	}
	path := getPath(context)
	value := context.Args().Get(1)

	SetValue(expected, path, value)
	actual, err := LoadConf([]string{file}, []string{})

	if err != nil {
		return path, nil, nil, err
	}
	if context.Bool("encrypt") {
		secretAgent, err := newSecretAgentFromCli(context)
		if err != nil {
			return path, nil, nil, fmt.Errorf("newSecretAgentFromCli failed: %v", err)
		}
		if err := secretAgent.DecryptPaths(actual, path); err != nil {
			return path, nil, nil, fmt.Errorf("DecryptPaths failed: %v", err)
		}
	}

	return path, expected, actual, nil
}

func testGetPath(t *testing.T, expected string, args []string, options []string) {
	path := getPath(NewTestContext("test", nil, globalFlags(), nil, args, options))
	if expected != path {
		t.Errorf("Get path failed: [%v] != [%v]", expected, path)
	}
}

func TestGetPath(t *testing.T) {
	var envVar = "CONFIG_PREFIX"
	defer func() {
		os.Unsetenv(envVar)
	}()

	testGetPath(t, "/", []string{""}, []string{})
	testGetPath(t, "/", []string{"/"}, []string{})
	testGetPath(t, "/foo", []string{"/foo"}, []string{})

	testGetPath(t, "/foo", []string{}, []string{"--prefix", "/foo"})
	testGetPath(t, "/foo", []string{"/"}, []string{"--prefix", "/foo"})
	testGetPath(t, "/foo/bar", []string{"/bar"}, []string{"--prefix", "/foo"})
	testGetPath(t, "/foo/bar", []string{"/bar"}, []string{"--prefix", "/foo/"})

	os.Setenv(envVar, "/foo")
	testGetPath(t, "/foo", []string{""}, []string{})
	testGetPath(t, "/foo", []string{"/"}, []string{})
	testGetPath(t, "/foo/bar", []string{"/bar"}, []string{})
}

func testGetValue(t *testing.T, args []string, options []string) {
	config, err := NewTestConfig()
	if err != nil {
		t.Error(err)
	}

	_, path, expected, actual, err := getGetvOutcome(config, args, options)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Getv %s unexpected result: %v != %v", path, expected, actual)
	}
}

func TestGetValue(t *testing.T) {
	testGetValue(t, []string{""}, []string{})
	testGetValue(t, []string{"/"}, []string{})
	testGetValue(t, []string{"/app"}, []string{})
	testGetValue(t, []string{"/app/db"}, []string{})
	testGetValue(t, []string{"/app/db/hostname"}, []string{})
	testGetValue(t, []string{"/app/db/hostname"}, []string{"--default", "INVALID_HOST"})
	testGetValue(t, []string{"INVALID_PATH"}, []string{})
	testGetValue(t, []string{"INVALID_PATH_WITH_DEFAULT"}, []string{"--default", "foo"})
	testGetValue(t, []string{"/app/db"}, []string{
		"--default", "foo",
		"--decrypt", "/username",
		"--decrypt", "/password",
	})
	testGetValue(t, []string{"/app/db/username"}, []string{"--decrypt", "/"})
}

func testGetValueWithTemplate(t *testing.T, name string, args, opts []string) {
	config, err := NewTestConfig()
	if err != nil {
		t.Error(err)
	}

	context, _, expected, actual, err := getGetvOutcome(config, args, opts)

	_, tmpl, err := getTemplate(context)
	if err != nil {
		t.Errorf("GetValueWithTemplate getTemplate %s failed and shouldn't have: %v", name, err)
	} else if tmpl == nil {
		t.Errorf("GetValueWithTemplate getTemplate %s expected result", name)
	}

	expectedString, err := tmpl.Execute(expected)
	if err != nil {
		t.Errorf("GetValueWithTemplate tmpl.Exectute %s failed and shouldn't have: %v", name, err)
	}
	// when templates are used, the template doesnt get processed
	// until marshaling...
	_, actualString, err := marshal(context, actual, nil)
	if err != nil {
		t.Errorf("GetValueWithTemplate marshal %s failed and shouldn't have: %v", name, err)
	}

	if expectedString != actualString {
		t.Errorf("GetValueWithTemplate %s invalid: [%v] != [%v]",
			name, expectedString, actualString)
	}
}

func TestGetValueWithTemplate(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	templateString := "{{ getv \"/username-plaintext\" }}:{{getv \"/password-plaintext\" }}"
	templateBytes := []byte(templateString)
	templateBase64 := base64.StdEncoding.EncodeToString(templateBytes)
	templateFile := filepath.Join(tempDir, "template")
	ioutil.WriteFile(templateFile, templateBytes, 0600)

	testGetValueWithTemplate(t, "getv template string",
		[]string{"/app/db"},
		[]string{"--template-string", templateString})
	testGetValueWithTemplate(t, "getv template base64",
		[]string{"/app/db"},
		[]string{"--template-base64", templateBase64})
	testGetValueWithTemplate(t, "getv template file",
		[]string{"/app/db"},
		[]string{"--template", templateFile})

	testGetValueWithTemplate(t, "cgetv template string",
		[]string{"/app/db"},
		[]string{
			"--template-string", "{{ cgetv \"/username\" }}:{{cgetv \"/password\" }}",
		})
}

func TestMarshal(t *testing.T) {
	var expected interface{}
	var actual interface{}

	context := NewTestGlobalContext()
	expected = "foo"
	context, actual, err := marshal(context, expected, nil)
	if context != context || actual != expected || err != nil {
		t.Errorf("Marshal string failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}

	expected = "2"
	context, actual, err = marshal(context, expected, nil)
	if context != context || actual != expected || err != nil {
		t.Errorf("Marshal int failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}

	expected, _ = UnmarshalYaml("a:\n  b: foo")
	context, marshaled, err := marshal(context, expected, nil)
	actual, _ = UnmarshalYaml(marshaled)
	if context != context || !reflect.DeepEqual(actual, expected) || err != nil {
		t.Errorf("Marshal map failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}

	expected, _ = UnmarshalYaml("a:\n- foo\n- bar")
	context, marshaled, err = marshal(context, expected, nil)
	actual, _ = UnmarshalYaml(marshaled)
	if context != context || !reflect.DeepEqual(actual, expected) || err != nil {
		t.Errorf("Marshal array failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}
}

func TestNewSecretAgentFromCli(t *testing.T) {
	var err error
	secretKeyringEnvVar := "SECRET_KEYRING"
	secretKeyringBase64EnvVar := "SECRET_KEYRING_BASE64"
	defer func() {
		// just in case
		os.Unsetenv(secretKeyringEnvVar)
		os.Unsetenv(secretKeyringBase64EnvVar)
	}()

	_, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{}))
	if err == nil {
		t.Errorf("New secret agent no options no env failed: [%v]", err)
	}

	secretAgent, err := newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil, []string{},
			[]string{"--secret-keyring", NewTestKeysFile()}))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from file failed: [%v]", err)
	}

	secretKeyring, err := ioutil.ReadFile(NewTestKeysFile())
	if err != nil {
		t.Errorf("New secret agent from base 64 read keys file failed: [%v]", err)
	}
	secretAgent, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil, []string{},
			[]string{"--secret-keyring-base64", base64.StdEncoding.EncodeToString(secretKeyring)}))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from base 64 failed: [%v]", err)
	}

	err = os.Setenv(secretKeyringEnvVar, NewTestKeysFile())
	if err != nil {
		t.Errorf("New secret agent from env set env failed: [%v]", err)
	}
	secretAgent, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{}))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from env failed: [%v]", err)
	}
	os.Unsetenv(secretKeyringEnvVar)

	err = os.Setenv(secretKeyringBase64EnvVar,
		base64.StdEncoding.EncodeToString(secretKeyring))
	if err != nil {
		t.Errorf("New secret agent from base 64 env set env failed: [%v]", err)
	}
	secretAgent, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{}))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from base 64 env failed: [%v]", err)
	}
	os.Unsetenv(secretKeyringEnvVar)
}

func TestCsetv(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	file := filepath.Join(tempDir, "config.yml")

	path, expected, actual, err := getCsetvOutcome(file, []string{"/a", "b"}, []string{})
	if err != nil {
		t.Errorf("Setv %s encrypt failed: %v", path, err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Setv %s encrypt failed: %v != %v", path, expected, actual)
	}
}

func TestSetv(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	file := filepath.Join(tempDir, "config.yml")

	context := NewTestSetvContext(file, []string{}, []string{})
	if err := setv(context); err == nil {
		t.Error("Setv no args should have failed")
	}

	context = NewTestSetvContext(file, []string{"/a"}, []string{})
	if err := setv(context); err == nil {
		t.Error("Setv one arg should have failed")
	}

	path, expected, actual, err := getSetvOutcome(file, []string{"/a", "b"}, []string{})
	if err != nil {
		t.Errorf("Setv %s failed: %v", path, err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Setv %s failed: %v != %v", path, expected, actual)
	}

	path, expected, actual, err = getSetvOutcome(file, []string{"/a", "b"}, []string{"--encrypt"})
	if err != nil {
		t.Errorf("Setv %s encrypt failed: %v", path, err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Setv %s encrypt failed: %v != %v", path, expected, actual)
	}
}
