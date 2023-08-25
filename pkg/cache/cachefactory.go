package cache

import (
	"github.com/falldio/Kache/pkg/config"
	log "github.com/sirupsen/logrus"
)

func NewDefaultCache(isHotCache bool) Cache {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic: %s", err)
		}
	}()
	switch config.Config.CacheStrategy {
	case CACHE_STRATEGY_FIFO:
		return newFIFOCache(config.Config.MaxCacheBytes)
	case CACHE_STRATEGY_LRU:
		return newLRUCache(config.Config.MaxCacheBytes)
	case CACHE_STRATEGY_LFU:
		return newLFUCache(config.Config.MaxCacheBytes)
	default:
		panic("unknown cache strategy: " + config.Config.CacheStrategy)
	}
}
