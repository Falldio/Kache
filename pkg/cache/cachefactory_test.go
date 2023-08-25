package cache

import (
	"reflect"
	"testing"

	"github.com/falldio/Kache/pkg/config"
)

func TestNewDefaultCache(t *testing.T) {
	config.Config.CacheStrategy = "fifo"
	c := NewDefaultCache(false)
	if reflect.TypeOf(c).String() != "*cache.FIFOCache" {
		t.Fatalf("expect *cache.FIFOCache, got %s", reflect.TypeOf(c).String())
	}
	config.Config.CacheStrategy = CACHE_STRATEGY_LRU
	c = NewDefaultCache(false)
	if reflect.TypeOf(c).String() != "*cache.LRUCache" {
		t.Fatalf("expect *cache.LRUCache, got %s", reflect.TypeOf(c).String())
	}
	config.Config.CacheStrategy = CACHE_STRATEGY_LFU
	c = NewDefaultCache(false)
	if reflect.TypeOf(c).String() != "*cache.LFUCache" {
		t.Fatalf("expect *cache.LFUCache, got %s", reflect.TypeOf(c).String())
	}
	config.Config.CacheStrategy = "unknown"
	c = NewDefaultCache(false)
	if c != nil {
		t.Fatalf("expect nil, got %s", reflect.TypeOf(c).String())
	}
}
