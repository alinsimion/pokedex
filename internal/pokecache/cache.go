package pokecache

import (
	"log/slog"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	cache map[string]cacheEntry
	lock  *sync.Mutex
}

func NewCache(interval time.Duration) Cache {

	tempCache := Cache{
		lock:  &sync.Mutex{},
		cache: make(map[string]cacheEntry),
	}
	go tempCache.reapLoop(interval)
	return tempCache
}

func (c Cache) Add(key string, val []byte) {

	tempCacheEntry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

	c.lock.Lock()

	c.cache[key] = tempCacheEntry

	c.lock.Unlock()
}

func (c Cache) Get(key string) ([]byte, bool) {

	if val, ok := c.cache[key]; ok {
		return val.val, ok
	}

	return []byte{}, false
}

func (c Cache) reapLoop(interval time.Duration) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		for key, elem := range c.cache {
			if time.Since(elem.createdAt) > interval {
				delete(c.cache, key)
				slog.Debug("removed", "key", key)
			}
		}
	}
}
