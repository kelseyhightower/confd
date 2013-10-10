package main

// A fileInfo describes a configuration file and is returned by fileStat.
type fileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode uint16
	Md5  string
}
