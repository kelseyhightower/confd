consul-kv
=========

*DEPRECATED* Please use [consul-api](https://github.com/armon/consul-api) instead.

This package provides the `consulkv` package which is a Key/Value
client for Consul. It supports all the commands as of Consul 0.1,
and has a very simple API.

Documentation
=============

The full documentation is available on [Godoc](http://godoc.org/github.com/armon/consul-kv)

Usage
=====

Below is an example of using the Consul KV client:

```go
client, _ := consulkv.NewClient(consulkv.DefaultConfig())

key := "foo"
value := []byte("test")
client.Put(key, value, 42)

meta, pair, err := client.Get(key)
fmt.Printf("Got %s (%d): %s", pair.Key, pair.Flags, pair.Value)
```

Development
===========

To run the tests, first run a consul server with API port 8500:

```
consul agent -data-dir /tmp/consul -server -bootstrap
```

The run tests:

```
go test
```
