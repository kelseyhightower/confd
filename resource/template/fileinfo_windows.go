// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package template

import (
	"os"
)

// fileInfo describes a configuration file and is returned by fileStat.
type fileInfo struct {
	Mode os.FileMode
	Md5  string
}
