package cache

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestGetFIFO(t *testing.T) {
	fifo := newFIFOCache(int64(0), nil)
	fifo.Set("k1", String("v1"))
	if v, ok := fifo.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("get key k1 failed, expect v1, got %v", v)
	}
	if _, ok := fifo.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestSetFIFO(t *testing.T) {
	// normal set
	fifo := newFIFOCache(int64(0), nil)
	fifo.Set("k1", String("v1"))
	fifo.Set("k2", String("v2"))
	if v, ok := fifo.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("get key k1 failed, expect v1, got %v", v)
	}
	if v, ok := fifo.Get("k2"); !ok || string(v.(String)) != "v2" {
		t.Fatalf("get key k2 failed, expect v2, got %v", v)
	}

	// update
	fifo.Set("k2", String("v3"))
	if v, ok := fifo.Get("k2"); !ok || string(v.(String)) != "v3" {
		t.Fatalf("get key k2 failed, expect v3, got %v", v)
	}

	// shrink
	keys := []string{}
	for i := 0; i < 10; i++ {
		keys = append(keys, fmt.Sprintf("%d+a", i))
	}
	fifo = newFIFOCache(int64(8), nil)
	for _, k := range keys {
		fifo.Set(k, String(k))
	}
	if v, ok := fifo.Get(keys[0]); ok {
		t.Fatalf("expect empty, got %v", v)
	}
	if v, ok := fifo.Get(keys[9]); !ok || string(v.(String)) != keys[9] {
		t.Fatalf("get key %s failed, expect %s, got %v", keys[9], keys[9], v)
	}
}

func TestRemoveFIFO(t *testing.T) {
	fifo := newFIFOCache(int64(0), nil)
	fifo.Set("key1", String("1234"))
	fifo.Remove("key1")
	if fifo.Has("key1") {
		t.Fatalf("fifo shouldn't have key1")
	}
}

func TestKeysFIFO(t *testing.T) {
	fifo := newFIFOCache(int64(0), nil)
	fifo.Set("k1", String("v1"))
	fifo.Set("k2", String("v2"))
	fifo.Set("k3", String("v3"))
	fifo.Set("k4", String("v4"))
	fifo.Set("k5", String("v5"))
	expect := []string{"k1", "k2", "k3", "k4", "k5"}

	keys := fifo.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("keys malperforming, expect: %v, got %v", expect, keys)
	}
}

func TestLenFIFO(t *testing.T) {
	fifo := newFIFOCache(int64(0), nil)
	sz := fifo.Len()
	if sz != 0 {
		t.Fatalf("fifo has wrong length, expect: 0, got: %d", sz)
	}
	fifo.Set("key1", String("1234"))
	sz = fifo.Len()
	if sz != 1 {
		t.Fatalf("fifo has wrong length, expect: 1, got: %d", sz)
	}
}

func TestHasFIFO(t *testing.T) {
	fifo := newFIFOCache(int64(0), nil)
	fifo.Set("key1", String("1234"))
	if !fifo.Has("key1") {
		t.Fatalf("fifo should have key1")
	}
	if fifo.Has("key2") {
		t.Fatalf("fifo shouldn't have key2")
	}
}

func TestShrinkFIFO(t *testing.T) {
	fifo := newFIFOCache(int64(0), nil)
	keys := []string{}
	for i := 0; i < 10; i++ {
		keys = append(keys, fmt.Sprintf("%d+a", i))
	}
	for _, k := range keys {
		fifo.Set(k, String(k))
	}
	if v, ok := fifo.Get(keys[0]); !ok {
		t.Fatalf("expect %v, got %v", keys[0], v)
	}
	fifo.Shrink()
	if v, ok := fifo.Get(keys[0]); ok {
		t.Fatalf("expect empty, got %v", v)
	}
}
