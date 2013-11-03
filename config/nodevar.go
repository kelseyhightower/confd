// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package config

import (
	"fmt"
)

// Nodes is a custom flag Var representing a list of etcd nodes. We use a custom
// Var to allow us to define more than one etcd node from the command line, and
// collect the results in a single value.
type Nodes []string

// String.
func (n *Nodes) String() string {
	return fmt.Sprintf("%d", *n)
}

// Set appends the node to the etcd node list.
func (n *Nodes) Set(node string) error {
	*n = append(*n, node)
	return nil
}
