package cache

import "sync"

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
	Set(key string, value Value)
	Remove(key string)
	Keys() []string
	Len() int
	Has(key string) bool
}

type baseCache struct {
	mu       sync.RWMutex
	maxBytes int64
	nbytes   int64 // current size
}