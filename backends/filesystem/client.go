package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	util "github.com/kelseyhightower/confd/util"
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
		fpath := toPath(key)
		stat, err := os.Stat(fpath)
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			// Walk subdirs
			flist, err := util.RecursiveFilesLookup(fpath, "*")
			if err != nil {
				return nil, err
			}

			klist := make([]string, 0)
			for _, v := range flist {
				klist = append(klist, toKey(v))
			}

			// Recursively get values for sub-keys (fles)
			recurseVars, err := c.GetValues(klist)
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
			b, err := ioutil.ReadFile(fpath)
			if err != nil {
				return nil, err
			}

			// Add to vars
			vars[key] = strings.TrimSpace(string(b))

		}
	}

	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
