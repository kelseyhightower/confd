package main

import (
	"os"
)

// isFileExist reports whether path exits.
func isFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}
