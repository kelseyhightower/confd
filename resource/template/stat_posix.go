// +build !windows

package template

import (
	"os"
	"syscall"
)


func uid(info os.FileInfo) uint32 {
	return info.Sys().(*syscall.Stat_t).Uid
}

func gid(info os.FileInfo) uint32 {
	return info.Sys().(*syscall.Stat_t).Gid
}

