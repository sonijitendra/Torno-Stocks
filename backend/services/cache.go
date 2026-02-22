package services

import (
	"sync"
	"time"
)

// CacheItem holds a cached value with expiry
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// MemoryCache is an in-memory TTL cache (Redis-ready abstraction)
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
	ttl   time.Duration
}

// NewMemoryCache creates a cache with the given TTL
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	c := &MemoryCache{items: make(map[string]CacheItem), ttl: ttl}
	go c.cleanup()
	return c
}

// Get returns the value if found and not expired
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	return item.Value, true
}

// Set stores a value with TTL
func (c *MemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	c.items[key] = CacheItem{Value: value, ExpiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.items {
			if now.After(v.ExpiresAt) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
