# memkv

Simple in memory k/v store.

[![Build Status](https://travis-ci.org/kelseyhightower/memkv.svg)](https://travis-ci.org/kelseyhightower/memkv) [![GoDoc](https://godoc.org/github.com/kelseyhightower/memkv?status.png)](https://godoc.org/github.com/kelseyhightower/memkv)

## Usage

```Go
package main

import (
	"fmt"

	"github.com/kelseyhightower/memkv"
)

func main() {
	s := memkv.New()
	s.Set("/myapp/database/username", "admin")
	s.Set("/myapp/database/password", "123456789")
	s.Set("/myapp/port", "80")

	// Get a specific node.
	node, ok := s.Get("/myapp/database/username")	
	if ok {
		fmt.Printf("Key: %s, Value: %s\n", node.Key, node.Value)
	}

	// Get all nodes where Key matches pattern.
	nodes, err := s.Glob("/myapp/*/*")
	if err == nil {
		for _, n := range nodes {
			fmt.Printf("Key: %s, Value: %s\n", n.Key, n.Value)
		}
	}
}	
```
