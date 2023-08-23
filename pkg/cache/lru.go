package cache

import (
	"container/list"
)

type LRUCache struct {
	baseCache
	// this list serves as the lru list, that is, the nodes are sorted according to their recent used time
	ll        *list.List
	items     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type lruEntry struct {
	key   string
	value Value
}

func newLRUCache(maxBytes int64, OnEvicted func(string, Value)) *LRUCache {
	return &LRUCache{
		baseCache: baseCache{maxBytes: maxBytes},
		ll:        list.New(),
		items:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

func (c *LRUCache) Get(key string) (value Value, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		kv := el.Value.(*lruEntry)
		return kv.value, true
	}
	return
}

func (c *LRUCache) removeOldest() {
	el := c.ll.Back()
	if el != nil {
		c.ll.Remove(el)
		kv := el.Value.(*lruEntry)
		delete(c.items, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Set(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		kv := el.Value.(*lruEntry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		el := c.ll.PushFront(&lruEntry{key, value})
		c.items[key] = el
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.removeOldest()
	}
}

func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *LRUCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	el, ok := c.items[key]
	if ok {
		c.ll.MoveToFront(el)
	}
	return ok
}

func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		kv := el.Value.(*lruEntry)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		delete(c.items, key)
		c.ll.Remove(el)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

var _ Cache = (*LRUCache)(nil)
