package kache

import (
	"fmt"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("callback failed, expect: %v, got: %v", expect, v)
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	g := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s does not exits", key)
		}))

	for k := range db {
		if _, err := g.Get(k); err != nil {
			t.Fatalf("failed to get value of %v", k)
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s missing", k)
		}
	}

	if view, err := g.Get("unknown"); err == nil {
		t.Fatalf("expect empty, get %s", view)
	}
}
