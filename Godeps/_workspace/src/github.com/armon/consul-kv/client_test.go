package consulkv

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"path"
	"testing"
	"time"
)

func testClient(t *testing.T) *Client {
	client, err := NewClient(DefaultConfig())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return client
}

func testKey() string {
	buf := make([]byte, 16)
	if _, err := crand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

func TestClientPutGetDelete(t *testing.T) {
	client := testClient(t)

	// Get a get without a key
	key := testKey()
	_, pair, err := client.Get(key)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pair != nil {
		t.Fatalf("unexpected value: %#v", pair)
	}

	// Put the key
	value := []byte("test")
	if err := client.Put(key, value, 42); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Get should work
	meta, pair, err := client.Get(key)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pair == nil {
		t.Fatalf("expected value: %#v", pair)
	}
	if !bytes.Equal(pair.Value, value) {
		t.Fatalf("unexpected value: %#v", pair)
	}
	if pair.Flags != 42 {
		t.Fatalf("unexpected value: %#v", pair)
	}
	if meta.ModifyIndex == 0 {
		t.Fatalf("unexpected value: %#v", meta)
	}

	// Delete
	if err := client.Delete(key); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Get should fail
	_, pair, err = client.Get(key)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pair != nil {
		t.Fatalf("unexpected value: %#v", pair)
	}
}

func TestClient_List_DeleteRecurse(t *testing.T) {
	client := testClient(t)

	// Generate some test keys
	prefix := testKey()
	var keys []string
	for i := 0; i < 100; i++ {
		keys = append(keys, path.Join(prefix, testKey()))
	}

	// Set values
	value := []byte("test")
	for _, key := range keys {
		if err := client.Put(key, value, 0); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// List the values
	meta, pairs, err := client.List(prefix)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(pairs) != len(keys) {
		t.Fatalf("got %d keys", len(pairs))
	}
	for _, pair := range pairs {
		if !bytes.Equal(pair.Value, value) {
			t.Fatalf("unexpected value: %#v", pair)
		}
	}
	if meta.ModifyIndex == 0 {
		t.Fatalf("unexpected value: %#v", meta)
	}

	// Delete all
	if err := client.DeleteTree(prefix); err != nil {
		t.Fatalf("err: %v", err)
	}

	// List the values
	_, pairs, err = client.List(prefix)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(pairs) != 0 {
		t.Fatalf("got %d keys", len(pairs))
	}
}

func TestClient_CAS(t *testing.T) {
	client := testClient(t)

	// Put the key
	key := testKey()
	value := []byte("test")
	if work, err := client.CAS(key, value, 0, 0); err != nil {
		t.Fatalf("err: %v", err)
	} else if !work {
		t.Fatalf("CAS failure")
	}

	// Get should work
	meta, pair, err := client.Get(key)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pair == nil {
		t.Fatalf("expected value: %#v", pair)
	}
	if meta.ModifyIndex == 0 {
		t.Fatalf("unexpected value: %#v", meta)
	}

	// CAS update with bad index
	newVal := []byte("foo")
	if work, err := client.CAS(key, newVal, 0, 1); err != nil {
		t.Fatalf("err: %v", err)
	} else if work {
		t.Fatalf("unexpected CAS")
	}

	// CAS update with valid index
	if work, err := client.CAS(key, newVal, 0, meta.ModifyIndex); err != nil {
		t.Fatalf("err: %v", err)
	} else if !work {
		t.Fatalf("unexpected CAS failure")
	}
}

func TestClient_WatchGet(t *testing.T) {
	client := testClient(t)

	// Get a get without a key
	key := testKey()
	meta, pair, err := client.Get(key)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pair != nil {
		t.Fatalf("unexpected value: %#v", pair)
	}
	if meta.ModifyIndex == 0 {
		t.Fatalf("unexpected value: %#v", meta)
	}

	// Put the key
	value := []byte("test")
	go func() {
		client := testClient(t)
		time.Sleep(100 * time.Millisecond)
		if err := client.Put(key, value, 42); err != nil {
			t.Fatalf("err: %v", err)
		}
	}()

	// Get should work
	meta2, pair, err := client.WatchGet(key, meta.ModifyIndex)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pair == nil {
		t.Fatalf("expected value: %#v", pair)
	}
	if !bytes.Equal(pair.Value, value) {
		t.Fatalf("unexpected value: %#v", pair)
	}
	if pair.Flags != 42 {
		t.Fatalf("unexpected value: %#v", pair)
	}
	if meta2.ModifyIndex <= meta.ModifyIndex {
		t.Fatalf("unexpected value: %#v", meta2)
	}
}

func TestClient_WatchList(t *testing.T) {
	client := testClient(t)

	// Get a get without a key
	prefix := testKey()
	key := path.Join(prefix, testKey())
	meta, pairs, err := client.List(prefix)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(pairs) != 0 {
		t.Fatalf("unexpected value: %#v", pairs)
	}
	if meta.ModifyIndex == 0 {
		t.Fatalf("unexpected value: %#v", meta)
	}

	// Put the key
	value := []byte("test")
	go func() {
		client := testClient(t)
		time.Sleep(100 * time.Millisecond)
		if err := client.Put(key, value, 42); err != nil {
			t.Fatalf("err: %v", err)
		}
	}()

	// Get should work
	meta2, pairs, err := client.WatchList(prefix, meta.ModifyIndex)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected value: %#v", pairs)
	}
	if !bytes.Equal(pairs[0].Value, value) {
		t.Fatalf("unexpected value: %#v", pairs)
	}
	if pairs[0].Flags != 42 {
		t.Fatalf("unexpected value: %#v", pairs)
	}
	if meta2.ModifyIndex <= meta.ModifyIndex {
		t.Fatalf("unexpected value: %#v", meta2)
	}

}
