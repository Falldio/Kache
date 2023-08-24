package kache

import (
	"fmt"
	"sync"

	"github.com/falldio/Kache/pkg/cache"
	"github.com/falldio/Kache/pkg/singleflight"
	log "github.com/sirupsen/logrus"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name   string
	getter Getter

	// mainCache stores kv that belongs to this peer according to the consistent hash algo
	mainCache cache.Cache

	// hotCache stores kv which is not allocated to this peer, but is so popular
	// that we would like to store it on every node, in order to avoid extra netwrok communication.
	hotCache cache.Cache

	cacheBytes int64 // total bytes limit of mainCache and hotCache

	peers PeerPicker

	// use singleflight.Group to make sure that each key
	// is only fetched once
	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:       name,
		getter:     getter,
		cacheBytes: cacheBytes,
		mainCache:  cache.NewDefaultCache(),
		hotCache:   cache.NewDefaultCache(),
		loader:     &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	v, cacheHit := g.lookupCache(key)
	if cacheHit {
		return v, nil
	}

	return g.load(key)
}

func (g *Group) lookupCache(key string) (value ByteView, ok bool) {
	if g.cacheBytes <= 0 {
		return
	}
	if v, ok := g.mainCache.Get(key); ok {
		return v.(ByteView), true
	}
	v, ok := g.hotCache.Get(key)
	if !ok {
		return
	}
	return v.(ByteView), ok
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than once")
	}
	g.peers = peers
}

// key is not in the local cache, we may have to ask other peers for help,
// or call local Getter method
func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[kache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	v, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bts: v}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bts, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{bts: cloneBytes(bts)}
	g.populateCache(key, value, &g.mainCache)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView, cache *cache.Cache) {
	if g.cacheBytes <= 0 {
		return
	}
	(*cache).Set(key, value)
	for {
		mainBytes := g.mainCache.Bytes()
		hotBytes := g.hotCache.Bytes()
		if mainBytes+hotBytes <= g.cacheBytes {
			return
		}
		victim := g.mainCache
		if hotBytes > mainBytes/16 {
			victim = g.hotCache
		}
		victim.Shrink()
	}
}
