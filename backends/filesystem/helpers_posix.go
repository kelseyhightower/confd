// +build !windows

package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

func toPath(key string) string {
	return filepath.FromSlash(key)
}

func toKey(path string) string {
	return filepath.ToSlash(path)
}