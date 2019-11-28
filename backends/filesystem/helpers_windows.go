package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

func toPath(key string) string {
	sep := string(os.PathSeparator)

	// Replace slashes with path seperator
	key = filepath.FromSlash(key)

	// Trim leading backslash, so \c\path\to\file becomes c\path\to\file
	key = strings.TrimPrefix(key, sep)

	// Replace first slash with colon-slash so c\path\to\file becomes c:\path\to\file
	key = strings.Replace(key, sep, ":"+sep, 1)

	return key
}

func toKey(path string) string {

	// Replace backslash with forward slash
	path = filepath.ToSlash(path)

	// Get rid of colon
	path = strings.Replace(path, ":", "", 1)

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
