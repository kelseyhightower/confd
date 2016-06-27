// +build windows
package template

import (
	"os"
)


func uid(info os.FileInfo) uint32 {
	return 0
}

func gid(info os.FileInfo) uint32 {
	return 0
}

