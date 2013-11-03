package template

import (
	"testing"
	"io/ioutil"
	"os"
)

var (
	fakeFile = "/this/shoud/not/exist"
)

func TestIsFileExist(t *testing.T) {
	result := IsFileExist(fakeFile)
	if result != false {
		t.Errorf("Expected IsFileExist(%s) to be false, got %v", fakeFile, result)
	}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(f.Name())
	result = IsFileExist(f.Name())
	if result != true {
		t.Errorf("Expected IsFileExist(%s) to be true, got %v", f.Name(), result)
	}
}
