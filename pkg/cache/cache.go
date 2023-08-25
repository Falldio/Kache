package cache

import (
	"sync"
	"time"
)

type cacheStrategy string

const (
	CACHE_STRATEGY_FIFO = "fifo"
	CACHE_STRATEGY_LRU  = "lru"
	CACHE_STRATEGY_LFU  = "lfu"
)

type Value interface {
	Len() int
}

type Cache interface {
	Get(key string) (Value, bool)
	Set(key string, value Value, ttl time.Duration)
	Remove(key string)
	Keys() []string
	Len() int
	Has(key string) bool
	Bytes() int64
	Shrink()
}

type baseCache struct {
	mu       sync.RWMutex
	maxBytes int64
	nbytes   int64 // current size
}

type cacheEntry struct {
	key   string
	value Value
	ttl   time.Time
}

func newBaseCache(maxBytes int64) baseCache {
	return baseCache{
		maxBytes: maxBytes,
	}
}

func (c *baseCache) Bytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nbytes
}
