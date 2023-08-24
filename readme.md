# Kache

Yet another distributed caching system.

Kache is a distributed caching system developed for learning purpose only. The general idea comes from [GeeCache](https://geektutu.com/post/geecache.html) and [GroupCache](https://github.com/golang/groupcache) (even the `readme` architecture comes from it). Some of the features are inspired by [PeanutCache](https://github.com/peanutzhen/peanutcache) and [gcache](https://github.com/bluele/gcache).

## Comparing to GeeCache

### Like GeeCache, Kache:

+ shard data by consistent hashing

### Unlike GeeCache, Kache:

+ use `gRPC` instead of `HTTP` as the communication protocol
+ support more caching strategies like lfu, fifo ...
+ support service discovery and registration by `etcd`

## TODO List

+ add hot cache support
+ add expiration mechanism
+ support data persistence
