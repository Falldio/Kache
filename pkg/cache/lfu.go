package cache

import (
	"container/list"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

type LFUCache struct {
	baseCache
	items   map[string]*list.Element
	freqMap map[int64]*list.List // freq -> freqList (LRU inside list)
	minFreq int64                // least freqency
	protect time.Duration        // protect time
}

type lfuEntry struct {
	key        string
	value      Value
	freq       int64
	insertTime time.Time
}

func newLFUCache(maxBytes int64) *LFUCache {
	return &LFUCache{
		baseCache: newBaseCache(maxBytes),
		items:     make(map[string]*list.Element),
		freqMap:   make(map[int64]*list.List),
		minFreq:   0,
		protect:   time.Millisecond * 5,
	}
}

func (c *LFUCache) Get(key string) (v Value, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		kv := el.Value.(*lfuEntry)
		c.updateFreq(el)
		return kv.value, true
	}
	log.Printf("key %s doesn't hit", key)
	return
}

func (c *LFUCache) Set(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.removeLeastFreqUsed()
	}
	if el, ok := c.items[key]; ok {
		kv := el.Value.(*lfuEntry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		el.Value = kv
		c.updateFreq(el)
	} else {
		if _, ok := c.freqMap[1]; !ok {
			c.freqMap[1] = list.New()
		}
		ev := &lfuEntry{key: key, freq: 1, value: value, insertTime: time.Now()}
		el := c.freqMap[1].PushFront(ev)
		c.items[key] = el
		c.minFreq = 1
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	log.Println("set key", key, "value", value, "nbytes", c.nbytes, "maxBytes", c.maxBytes)
}

func (c *LFUCache) updateFreq(el *list.Element) {
	kv := el.Value.(*lfuEntry)
	c.freqMap[kv.freq].Remove(el)
	kv.freq++
	if l, ok := c.freqMap[kv.freq]; ok {
		c.items[kv.key] = l.PushFront(kv)
	} else {
		c.freqMap[kv.freq] = list.New()
		c.items[kv.key] = c.freqMap[kv.freq].PushFront(kv)
	}
	if l, ok := c.freqMap[c.minFreq]; !ok || l.Len() == 0 {
		delete(c.freqMap, c.minFreq)
		c.minFreq++
	}
}

func (c *LFUCache) removeLeastFreqUsed() {
	el := c.freqMap[c.minFreq].Back()
	for el != nil && el.Value.(*lfuEntry).insertTime.Add(c.protect).After(time.Now()) {
		el = el.Prev()
	}
	if el == nil {
		el = c.freqMap[c.minFreq].Back()
	}
	c.remove(el)
}

func (c *LFUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el := c.items[key]
	c.remove(el)
}

func (c *LFUCache) remove(el *list.Element) {
	if el != nil {
		kv := el.Value.(*lfuEntry)
		log.Println("removing key", kv.key, "value", kv.value, "nbytes", c.nbytes, "maxBytes", c.maxBytes)
		c.freqMap[c.minFreq].Remove(el)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		delete(c.items, kv.key)
		if c.freqMap[c.minFreq].Len() == 0 {
			delete(c.freqMap, c.minFreq)
			min := int64(math.MaxInt64)
			for f := range c.freqMap {
				if f < min {
					min = f
				}
			}
			c.minFreq = min
		}
	}
}

func (c *LFUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

func (c *LFUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *LFUCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	el, ok := c.items[key]
	if ok {
		c.updateFreq(el)
	}
	return ok
}

func (c *LFUCache) Shrink() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removeLeastFreqUsed()
}

var _ Cache = (*LFUCache)(nil)
