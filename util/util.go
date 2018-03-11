package template

import (
   "fmt"
   "os"
   "path"
   "path/filepath"
   "github.com/kelseyhightower/confd/log"
)

// fileInfo describes a configuration file and is returned by fileStat.
type FileInfo struct {
   Uid  uint32
   Gid  uint32
   Mode os.FileMode
   Md5  string
}

func AppendPrefix(prefix string, keys []string) []string {
   s := make([]string, len(keys))
   for i, k := range keys {
      s[i] = path.Join(prefix, k)
   }
   return s
}

// isFileExist reports whether path exits.
func IsFileExist(fpath string) bool {
   if _, err := os.Stat(fpath); os.IsNotExist(err) {
      return false
   }
   return true
}

// IsConfigChanged reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func IsConfigChanged(src, dest string) (bool, error) {
   if !IsFileExist(dest) {
      return true, nil
   }
   d, err := filestat(dest)
   if err != nil {
      return true, err
   }
   s, err := filestat(src)
   if err != nil {
      return true, err
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
      return true, nil
   }
   return false, nil
}




func IsDirectory(path string) (bool, error) {
   f, err := os.Stat(path)
   if err != nil {
      return false, err
   }
   switch mode := f.Mode(); {
   case mode.IsDir():
      return true, nil
   case mode.IsRegular():
      return false, nil
   }
   return false, nil
}

func Lookup(root string, pattern string) ([]string, []string, error) {
   var files []string
   var dirs []string
   isDir, err := IsDirectory(root)
   if err != nil {
      return nil, nil, err
   }
   if isDir {
      err := filepath.Walk(root, func(root string, f os.FileInfo, err error) error {
         match, err := filepath.Match(pattern, f.Name())
         if err != nil {
            return err
         }
         if match {
            isDir, err := IsDirectory(root)
            if err != nil {
               return err
            }
            if isDir {
               dirs = append(dirs, root)
            } else {
               files = append(files, root)
            }
         }
         return nil
      })
      if err != nil {
         return nil, nil, err
      }
   } else {
      files = append(files, root)
   }
   return files, dirs, nil
}

