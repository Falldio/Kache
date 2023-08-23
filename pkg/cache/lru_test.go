package cache

import (
	"reflect"
	"sort"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGetLRU(t *testing.T) {
	lru := newLRUCache(int64(0), nil)
	lru.Set("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldestLRU(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := newLRUCache(int64(cap), nil)
	lru.Set(k1, String(v1))
	lru.Set(k2, String(v2))
	lru.Set(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.ll.Len() != 2 {
		t.Fatalf("RemoveOlderst key1 failed")
	}
}

func TestOnEnvictedLRU(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := newLRUCache(int64(10), callback)
	lru.Set("key1", String("123456"))
	lru.Set("k2", String("v2"))
	lru.Set("k3", String("v3"))
	lru.Set("k4", String("v4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect: %v, got: %v", expect, keys)
	}
}

func TestKeysLRU(t *testing.T) {
	lru := newLRUCache(int64(0), nil)
	lru.Set("k1", String("v1"))
	lru.Set("k2", String("v2"))
	lru.Set("k3", String("v3"))
	lru.Set("k4", String("v4"))
	lru.Set("k5", String("v5"))
	expect := []string{"k1", "k2", "k3", "k4", "k5"}

	keys := lru.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("keys malperforming, expect: %v, got %v", expect, keys)
	}
}

func TestHasLRU(t *testing.T) {
	lru := newLRUCache(int64(0), nil)
	lru.Set("key1", String("1234"))
	if !lru.Has("key1") {
		t.Fatalf("lru should have key1")
	}
	if lru.Has("key2") {
		t.Fatalf("lru shouldn't have key2")
	}
}

func TestRemoveLRU(t *testing.T) {
	lru := newLRUCache(int64(0), nil)
	lru.Set("key1", String("1234"))
	lru.Remove("key1")
	if lru.Has("key1") {
		t.Fatalf("lru shouldn't have key1")
	}
}

func TestLenLRU(t *testing.T) {
	lru := newLRUCache(int64(0), nil)
	sz := lru.Len()
	if sz != 0 {
		t.Fatalf("lru has wrong length, expect: 0, got: %d", sz)
	}
	lru.Set("key1", String("1234"))
	sz = lru.Len()
	if sz != 1 {
		t.Fatalf("lru has wrong length, expect: 1, got: %d", sz)
	}
}
