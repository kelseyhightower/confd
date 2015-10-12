package template

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"

	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/util"
)

// fileInfo describes a configuration file and is returned by fileStat.
type fileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode os.FileMode
	Md5  string
}

func appendPrefix(prefix string, keys []string) []string {
	s := make([]string, len(keys))
	for i, k := range keys {
		s[i] = path.Join(prefix, k)
	}
	return s
}

// fileStat return a fileInfo describing the named file.
func fileStat(name string) (fi fileInfo, err error) {
	if util.IsFileExist(name) {
		f, err := os.Open(name)
		if err != nil {
			return fi, err
		}
		defer f.Close()
		stats, _ := f.Stat()
		fi.Uid = stats.Sys().(*syscall.Stat_t).Uid
		fi.Gid = stats.Sys().(*syscall.Stat_t).Gid
		fi.Mode = stats.Mode()
		h := md5.New()
		io.Copy(h, f)
		fi.Md5 = fmt.Sprintf("%x", h.Sum(nil))
		return fi, nil
	} else {
		return fi, errors.New("File not found")
	}
}

// sameConfig reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func sameConfig(src, dest string) (bool, error) {
	if !util.IsFileExist(dest) {
		return false, nil
	}
	d, err := fileStat(dest)
	if err != nil {
		return false, err
	}
	s, err := fileStat(src)
	if err != nil {
		return false, err
	}
	if d.Uid != s.Uid {
		log.Info(fmt.Sprintf("%s has UID %d should be %d", dest, d.Uid, s.Uid))
	}
	if d.Gid != s.Gid {
		log.Info(fmt.Sprintf("%s has GID %d should be %d", dest, d.Gid, s.Gid))
	}
	if d.Mode != s.Mode {
		log.Info(fmt.Sprintf("%s has mode %s should be %s", dest, os.FileMode(d.Mode), os.FileMode(s.Mode)))
	}
	if d.Md5 != s.Md5 {
		log.Info(fmt.Sprintf("%s has md5sum %s should be %s", dest, d.Md5, s.Md5))
	}
	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Md5 != s.Md5 {
		return false, nil
	}
	return true, nil
}

