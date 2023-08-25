package cache

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGetLRU(t *testing.T) {
	lru := newLRUCache(int64(0))
	lru.Set("key1", String("1234"), 0)
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestSetLRU(t *testing.T) {
	// normal set
	lru := newLRUCache(int64(0))
	lru.Set("key1", String("1234"), 0)
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	// update
	lru.Set("key1", String("5678"), 0)
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "5678" {
		t.Fatalf("cache hit key1=5678 failed")
	}
}

func TestRemoveOldestLRU(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := newLRUCache(int64(cap))
	lru.Set(k1, String(v1), 0)
	lru.Set(k2, String(v2), 0)
	lru.Set(k3, String(v3), 0)

	if _, ok := lru.Get("key1"); ok || lru.ll.Len() != 2 {
		t.Fatalf("RemoveOlderst key1 failed")
	}
}

func TestKeysLRU(t *testing.T) {
	lru := newLRUCache(int64(0))
	lru.Set("k1", String("v1"), 0)
	lru.Set("k2", String("v2"), 0)
	lru.Set("k3", String("v3"), 0)
	lru.Set("k4", String("v4"), 0)
	lru.Set("k5", String("v5"), 0)
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
	lru := newLRUCache(int64(0))
	lru.Set("key1", String("1234"), 0)
	if !lru.Has("key1") {
		t.Fatalf("lru should have key1")
	}
	if lru.Has("key2") {
		t.Fatalf("lru shouldn't have key2")
	}
}

func TestRemoveLRU(t *testing.T) {
	lru := newLRUCache(int64(0))
	lru.Set("key1", String("1234"), 0)
	lru.Remove("key1")
	if lru.Has("key1") {
		t.Fatalf("lru shouldn't have key1")
	}
}

func TestLenLRU(t *testing.T) {
	lru := newLRUCache(int64(0))
	sz := lru.Len()
	if sz != 0 {
		t.Fatalf("lru has wrong length, expect: 0, got: %d", sz)
	}
	lru.Set("key1", String("1234"), 0)
	sz = lru.Len()
	if sz != 1 {
		t.Fatalf("lru has wrong length, expect: 1, got: %d", sz)
	}
}

func TestShrinkLRU(t *testing.T) {
	lru := newLRUCache(int64(0))
	lru.Set("k1", String("v1"), 0)
	lru.Set("k2", String("v2"), 0)
	// has function will move k1 to the front of the list,
	// so k2 will be removed when shrink is called
	if !lru.Has("k1") {
		t.Fatalf("lru should have k1")
	}
	lru.Shrink()
	if lru.Has("k2") {
		t.Fatalf("lru shouldn't have k1")
	}
}

func TestExpireLRU(t *testing.T) {
	lru := newLRUCache(int64(0))
	lru.Set("k1", String("v1"), time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)
	if lru.Has("k1") {
		t.Fatalf("lru should not have k1")
	}
	if _, ok := lru.Get("k1"); ok {
		t.Fatalf("lru should not have k1")
	}
	// update ttl
	lru.Set("k1", String("v1"), time.Millisecond*20)
	lru.Set("k1", String("v1"), time.Millisecond*10)
	if v, ok := lru.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("get key k1 failed, expect v1, got %v", v)
	}
	time.Sleep(time.Millisecond * 10)
	if lru.Has("k1") {
		t.Fatalf("lru should not have k1")
	}

	// remove on get
	lru.Set("k1", String("v1"), time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)
	if _, ok := lru.Get("k1"); ok {
		t.Fatalf("lru should not have k1")
	}
}
