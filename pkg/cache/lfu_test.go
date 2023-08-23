package cache

import (
	"reflect"
	"sort"
	"testing"
)

func TestGetLFU(t *testing.T) {
	lfu := newLFUCache(int64(0), nil)
	lfu.Set("k1", String("v1"))
	if v, ok := lfu.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("get key k1 failed, expect v1, got %v", v)
	}
	if _, ok := lfu.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveLFU(t *testing.T) {
	lfu := newLFUCache(int64(0), nil)
	lfu.Set("key1", String("1234"))
	lfu.Remove("key1")
	if lfu.Has("key1") {
		t.Fatalf("lfu shouldn't have key1")
	}
}

func TestKeysLFU(t *testing.T) {
	lfu := newLFUCache(int64(0), nil)
	lfu.Set("k1", String("v1"))
	lfu.Set("k2", String("v2"))
	lfu.Set("k3", String("v3"))
	lfu.Set("k4", String("v4"))
	lfu.Set("k5", String("v5"))
	expect := []string{"k1", "k2", "k3", "k4", "k5"}

	keys := lfu.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("keys malperforming, expect: %v, got %v", expect, keys)
	}
}

func TestLenLFU(t *testing.T) {
	lfu := newLFUCache(int64(0), nil)
	sz := lfu.Len()
	if sz != 0 {
		t.Fatalf("lfu has wrong length, expect: 0, got: %d", sz)
	}
	lfu.Set("key1", String("1234"))
	sz = lfu.Len()
	if sz != 1 {
		t.Fatalf("lfu has wrong length, expect: 1, got: %d", sz)
	}
}

func TestHasLFU(t *testing.T) {
	lfu := newLFUCache(int64(0), nil)
	lfu.Set("key1", String("1234"))
	if !lfu.Has("key1") {
		t.Fatalf("lfu should have key1")
	}
	if lfu.Has("key2") {
		t.Fatalf("lfu shouldn't have key2")
	}
}
