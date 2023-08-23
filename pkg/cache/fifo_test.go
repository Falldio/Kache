package cache

import (
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
