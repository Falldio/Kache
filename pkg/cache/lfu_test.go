package cache

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestGetLFU(t *testing.T) {
	lfu := newLFUCache(int64(0))
	lfu.Set("k1", String("v1"), 0)
	if v, ok := lfu.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("get key k1 failed, expect v1, got %v", v)
	}
	if _, ok := lfu.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestSetLFU(t *testing.T) {
	lfu := newLFUCache(int64(0))
	lfu.Set("k1", String("v1"), 0)
	lfu.Set("k2", String("v2"), 0)
	if v, ok := lfu.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("get key k1 failed, expect v1, got %v", v)
	}
	if v, ok := lfu.Get("k2"); !ok || string(v.(String)) != "v2" {
		t.Fatalf("get key k2 failed, expect v2, got %v", v)
	}
	// update
	lfu.Set("k2", String("v3"), 0)
	if v, ok := lfu.Get("k2"); !ok || string(v.(String)) != "v3" {
		t.Fatalf("get key k2 failed, expect v3, got %v", v)
	}
	// shrink
	keys := []string{}
	for i := 0; i < 10; i++ {
		keys = append(keys, fmt.Sprintf("%d+a", i))
	}
	lfu = newLFUCache(int64(36))
	for i, k := range keys {
		lfu.Set(k, String(k), 0)
		for j := 0; j < i; j++ {
			lfu.Get(k)
		}
		time.Sleep(time.Millisecond * 10)
	}
	if v, ok := lfu.Get(keys[0]); ok {
		t.Fatalf("expect empty, got %v", v)
	}
	if v, ok := lfu.Get(keys[9]); !ok || string(v.(String)) != keys[9] {
		t.Fatalf("get key %s failed, expect %s, got %v", keys[9], keys[9], v)
	}

}

func TestRemoveLFU(t *testing.T) {
	lfu := newLFUCache(int64(0))
	lfu.Set("key1", String("1234"), 0)
	lfu.Remove("key1")
	if lfu.Has("key1") {
		t.Fatalf("lfu shouldn't have key1")
	}
}

func TestKeysLFU(t *testing.T) {
	lfu := newLFUCache(int64(0))
	lfu.Set("k1", String("v1"), 0)
	lfu.Set("k2", String("v2"), 0)
	lfu.Set("k3", String("v3"), 0)
	lfu.Set("k4", String("v4"), 0)
	lfu.Set("k5", String("v5"), 0)
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
	lfu := newLFUCache(int64(0))
	sz := lfu.Len()
	if sz != 0 {
		t.Fatalf("lfu has wrong length, expect: 0, got: %d", sz)
	}
	lfu.Set("key1", String("1234"), 0)
	sz = lfu.Len()
	if sz != 1 {
		t.Fatalf("lfu has wrong length, expect: 1, got: %d", sz)
	}
}

func TestHasLFU(t *testing.T) {
	lfu := newLFUCache(int64(0))
	lfu.Set("key1", String("1234"), 0)
	if !lfu.Has("key1") {
		t.Fatalf("lfu should have key1")
	}
	if lfu.Has("key2") {
		t.Fatalf("lfu shouldn't have key2")
	}
}

func TestShrinkLFU(t *testing.T) {
	lfu := newLFUCache(int64(0))
	lfu.Set("k1", String("v1"), 0)
	if !lfu.Has("k1") {
		t.Fatalf("lfu should have key1")
	}
	lfu.Shrink()
	if lfu.Has("k1") {
		t.Fatalf("lfu shouldn't have key1")
	}
}

func TestExpireLFU(t *testing.T) {
	lfu := newLFUCache(int64(0))
	lfu.Set("k1", String("v1"), time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)
	if lfu.Has("k1") {
		t.Fatalf("lfu shouldn't have key1")
	}
	if v, ok := lfu.Get("k1"); ok {
		t.Fatalf("lfu shouldn't have key1, got %v", v)
	}
	// update ttl
	lfu.Set("k1", String("v1"), time.Millisecond*20)
	lfu.Set("k1", String("v1"), time.Millisecond*10)
	if v, ok := lfu.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("get key k1 failed, expect v1, got %v", v)
	}
	time.Sleep(time.Millisecond * 10)
	if lfu.Has("k1") {
		t.Fatalf("lfu shouldn't have key1")
	}

	// remove on get
	lfu.Set("k1", String("v1"), time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)
	if _, ok := lfu.Get("k1"); ok {
		t.Fatalf("lfu shouldn't have key1")
	}
}
