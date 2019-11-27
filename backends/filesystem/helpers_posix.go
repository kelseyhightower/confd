// +build !windows

package filesystem

import (
	"path/filepath"
)

func toPath(key string) string {
	return filepath.FromSlash(key)
}

func toKey(path string) string {
	return filepath.ToSlash(path)
}
