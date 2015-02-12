package fs

import (
	"os"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/kelseyhightower/confd/log"
)

type Client struct {
	rootPath    string
	maxFileSize int64
}

func NewFsClient(rootPath string, maxFileSize int) (*Client, error) {
	return &Client{rootPath, int64(maxFileSize)}, nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		keyPath := path.Join(c.rootPath, key)
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			return vars, err
		}

		walk := func(fpath string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !fi.IsDir() {
				if (fi.Size() > c.maxFileSize) {
					log.Error("Maximum file size (" + strconv.Itoa(int(c.maxFileSize)) + ") exceeded for: " + fi.Name())
					return nil
				}
				v, err := ioutil.ReadFile(fpath)
				if err != nil {
					log.Error("Unable to read file: " + fpath)
					return nil
				}
				relFilePath := strings.TrimPrefix(fpath, c.rootPath)
				vars[relFilePath] = string(v)
			}
			return nil
		}

		if err := filepath.Walk(keyPath, walk); err != nil {
			return vars, err
		}
	}
	return vars, nil
}

// WatchPrefix is not yet implemented.
// A good start is fsnotify
// URL https://github.com/go-fsnotify/fsnotify
func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
