package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Client provides a shell for the filesystem client
type Client struct {
	MaxFileSize int64
}

// NewFilesystemClient returns a new client
func NewFileSystemClient(max int64) (*Client, error) {
	return &Client{MaxFileSize: max}, nil
}

// GetValues queries the filesystem for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		stat, err := os.Stat(key)
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			fileList := make([]string, 0)

			// Walk subdirs
			err := filepath.Walk(key, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					fileList = append(fileList, info.Name())
				}
				return nil
			})
			if err != nil {
				return nil, err
			}

			// Recursively get values for sub-keys (files)
			recurseVars, err := c.GetValues(fileList)
			if err != nil {
				return nil, err
			}

			// Add to our vars map
			for k, v := range recurseVars {
				vars[k] = v
			}
		} else {
			if stat.Size() > c.MaxFileSize {
				return nil, fmt.Errorf("size of %s too large", key)
			}

			// Read contents of file
			b, err := ioutil.ReadFile(key)
			if err != nil {
				return nil, err
			}

			// Add to vars
			vars[key] = string(b)
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
