package cache

import (
	"container/list"

	log "github.com/sirupsen/logrus"
)

type FIFOCache struct {
	baseCache
	items map[string]*list.Element
	ll    *list.List // fifo list
}

type fifoEntry struct {
	key   string
	value Value
}

func newFIFOCache(maxBytes int64) *FIFOCache {
	return &FIFOCache{
		baseCache: newBaseCache(maxBytes),
		items:     make(map[string]*list.Element),
		ll:        list.New(),
	}
}

func (c *FIFOCache) Get(key string) (value Value, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.items[key]; ok {
		return v.Value.(*fifoEntry).value, true
	}
	log.Printf("cache miss key: %s", key)
	return
}

func (c *FIFOCache) Set(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.items[key]; ok {
		c.nbytes += int64(value.Len()) - int64(v.Value.(*fifoEntry).value.Len())
		c.items[key].Value = &fifoEntry{key: key, value: value}
	} else {
		c.nbytes += int64(len(key)) + int64(value.Len())
		c.items[key] = c.ll.PushFront(&fifoEntry{key: key, value: value})
	}
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.shrink()
	}
}

func (c *FIFOCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.ll.Remove(el)
		kv := el.Value.(*fifoEntry)
		delete(c.items, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
	}
}

func (c *FIFOCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

func (c *FIFOCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *FIFOCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.items[key]
	return ok
}

func (c *FIFOCache) Shrink() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shrink()
}

func (c *FIFOCache) shrink() {
	el := c.ll.Back()
	c.ll.Remove(el)
	kv := el.Value.(*fifoEntry)
	delete(c.items, kv.key)
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
}

var _ Cache = (*FIFOCache)(nil)
