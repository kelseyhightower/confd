package template

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/kelseyhightower/confd/logging"
)

// createRecursiveDirs is a helper function which creates temporary directorie
// has sub directories, and files with different extionname which will be used
// fo function findrecursiveFindFiles's test case.The result looks like:
// ├── root.other1
// ├── root.toml
// ├── subDir1
// │   ├── sub1.other
// │   ├── sub1.toml
// │   └── sub12.toml
// └── subDir2
//			├── sub2.other
//			├── sub2.toml
//			├── sub22.toml
//			└── subSubDir
//					├── subsub.other
//					├── subsub.toml
//					└── subsub2.toml
func createRecursiveDirs() (rootDir string, err error) {
	mod := os.FileMode(0755)
	flag := os.O_RDWR | os.O_CREATE | os.O_EXCL
	rootDir, err = ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(rootDir+"/root.toml", flag, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(rootDir+"/root.other1", flag, mod)
	if err != nil {
		return "", err
	}
	subDir := filepath.Join(rootDir, "subDir1")
	err = os.Mkdir(subDir, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subDir+"/sub1.toml", flag, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subDir+"/sub12.toml", flag, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subDir+"/sub1.other", flag, mod)
	if err != nil {
		return "", err
	}
	subDir2 := filepath.Join(rootDir, "subDir2")
	err = os.Mkdir(subDir2, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subDir2+"/sub2.toml", flag, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subDir2+"/sub22.toml", flag, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subDir2+"/sub2.other", flag, mod)
	if err != nil {
		return "", err
	}
	subSubDir := filepath.Join(subDir2, "subSubDir")
	err = os.Mkdir(subSubDir, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subSubDir+"/subsub.toml", flag, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subSubDir+"/subsub2.toml", flag, mod)
	if err != nil {
		return "", err
	}
	_, err = os.OpenFile(subSubDir+"/subsub.other", flag, mod)
	if err != nil {
		return "", err
	}
	return
}

func TestRecursiveFindFiles(t *testing.T) {
	logging.SetLevel("warn")
	// Setup temporary directories
	rootDir, err := createRecursiveDirs()
	if err != nil {
		t.Errorf("Failed to create temp dirs: %s", err.Error())
	}
	defer os.RemoveAll(rootDir)
	files, err := recursiveFindFiles(rootDir, "*toml")
	if err != nil {
		t.Errorf("Failed to run recursiveFindFiles, got error: " + err.Error())
	}
	sort.Strings(files)
	exceptedFiles := []string{
		rootDir + "/" + "root.toml",
		rootDir + "/subDir1/" + "sub1.toml",
		rootDir + "/subDir1/" + "sub12.toml",
		rootDir + "/subDir2/" + "sub2.toml",
		rootDir + "/subDir2/" + "sub22.toml",
		rootDir + "/subDir2/subSubDir/" + "subsub.toml",
		rootDir + "/subDir2/subSubDir/" + "subsub2.toml",
	}
	for i, f := range exceptedFiles {
		if f != files[i] {
			t.FailNow()
		}
	}
}
