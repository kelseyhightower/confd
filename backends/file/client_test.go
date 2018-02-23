package file

import (
	"reflect"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const data = "" +
	"a: Baz\n" +
	"b:\n" +
	" c: Foo\n" +
	" d: 1\n" +
	" e: 1.1\n" +
	" f:\n" +
	" - Bar\n" +
	" - 1\n" +
	" - 1.1"
var expected = map[string]string{
	"/a": "Baz",
	"/b/c": "Foo",
	"/b/d": "1",
	"/b/e": "1.1",
	"/b/f/Bar": "",
	"/b/f/1": "",
	"/b/f/1.1": ""}

func getValuesFrom(data string, keys ...string) (map[string]string, error) {
	tempDir := os.TempDir()
	defer func() {
		os.RemoveAll(tempDir)
	}()

	file := path.Join(tempDir, "data.yml")
	ioutil.WriteFile(file, []byte(data), 0700)

	client, err := NewFileClient(file)
	if err != nil {
	    return nil, err
	}

	return client.GetValues(keys)
}

func TestGetValues(t *testing.T) {
	values, err := getValuesFrom(data, "/") 
	if err != nil {
	    t.Errorf("Failed to get values: %v", err)
	}

	if !reflect.DeepEqual(expected, values) {
	    t.Errorf("Failed get values: [%v] != [%v]", expected, values)
	}
}