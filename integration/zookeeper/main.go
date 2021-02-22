package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	zk "github.com/samuel/go-zookeeper/zk"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func zk_write(k string, v string, c *zk.Conn) {
	e, stat, err := c.Exists(k)
	check(err)
	if e {
		stat, err = c.Set(k, []byte(v), stat.Version)
		check(err)
	} else {
		string, err := c.Create(k, []byte(v), int32(0), zk.WorldACL(zk.PermAll))
		if string == "" {
			check(err)
		}
	}
}

func parsejson(prefix string, x interface{}, c *zk.Conn) {

	switch t := x.(type) {
	case map[string]interface{}:
		for k, v := range t {
			if prefix != "" {
				zk_write(prefix, "", c)
			}
			parsejson(prefix+"/"+k, v, c)
		}
	case []interface{}:
		for i, v := range t {
			parsejson(prefix+"["+strconv.Itoa(i)+"]", v, c)
		}
	case string:
		zk_write(prefix, t, c)
		fmt.Printf("%s = %q\n", prefix, t)
	default:
		fmt.Printf("Unhandled: %T\n", t)
	}
}

func main() {
	var pj interface{}
	dat, err := ioutil.ReadFile("test.json")
	check(err)
	err = json.Unmarshal(dat, &pj)
	check(err)
	c, _, err := zk.Connect([]string{os.Getenv("ZOOKEEPER_NODE")}, time.Second) //*10)
	check(err)
	parsejson("", pj, c)
}
