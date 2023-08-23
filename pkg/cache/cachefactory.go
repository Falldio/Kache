package cache

import "github.com/falldio/Kache/pkg/config"

func NewDefaultCache() Cache {
	switch config.Config.CacheStrategy {
	case CACHE_STRATEGY_LRU:
		return newLRUCache(config.Config.MaxCacheBytes, nil)
	case CACHE_STRATEGY_LFU:
		return newLFUCache(config.Config.MaxCacheBytes, nil)
	default:
		panic("unknown cache strategy: " + config.Config.CacheStrategy)
	}
}
