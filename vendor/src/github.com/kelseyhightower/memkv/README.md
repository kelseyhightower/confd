# memkv

Simple in memory k/v store.

[![Build Status](https://travis-ci.org/kelseyhightower/memkv.svg)](https://travis-ci.org/kelseyhightower/memkv) [![GoDoc](https://godoc.org/github.com/kelseyhightower/memkv?status.png)](https://godoc.org/github.com/kelseyhightower/memkv)

## Usage

```Go
package main

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/memkv"
)

func main() {
	s := memkv.New()
	s.Set("/myapp/database/username", "admin")
	s.Set("/myapp/database/password", "123456789")
	s.Set("/myapp/port", "80")
	kv, err := s.Get("/myapp/database/username")	
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Key: %s, Value: %s\n", kv.Key, kv.Value)
	ks, err := s.GetAll("/myapp/*/*")
	if err == nil {
		for _, kv := range ks {
			fmt.Printf("Key: %s, Value: %s\n", kv.Key, kv.Value)
		}
	}
}
```

---

```
Key: /myapp/database/username, Value: admin
Key: /myapp/database/password, Value: 123456789
Key: /myapp/database/username, Value: admin
```
