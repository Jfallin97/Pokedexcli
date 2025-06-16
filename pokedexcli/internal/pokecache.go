package internal

import (
	"sync"
	"time"
)

type Cache struct {
	m        map[string]cacheEntry
	mu       sync.Mutex
	interval time.Duration
}

func (c *Cache) Add(key string, val []byte) {
	cache_Entry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = cache_Entry

}

func (c *Cache) Get(key string) ([]byte, bool) {
	var data []byte

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.m[key]

	if ok {
		data = entry.val
		return data, true
	}
	return data, false
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	for {
		<-ticker.C

		c.mu.Lock()
		for key, entry := range c.m {
			if time.Now().After(entry.createdAt.Add(c.interval)) {
				delete(c.m, key)
			}
		}
		c.mu.Unlock()
	}
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		m:        make(map[string]cacheEntry),
		mu:       sync.Mutex{},
		interval: interval,
	}

	go cache.reapLoop()

	return cache
}
