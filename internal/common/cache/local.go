// Package cache provides caching utilities for the Travel Booking platform.
//
// CHK054: Local in-memory cache layer (Level 3 of 5-level cache architecture).
// Short TTL local cache sits in front of Redis to reduce network round-trips
// for hot data like product listings and configuration.
package cache

import (
	"sync"
	"time"
)

// LocalCache is a thread-safe in-memory cache with TTL support.
// Used as Level 3 cache (browser → CDN → local → Redis → DB).
type LocalCache struct {
	items sync.Map
	ttl   time.Duration
}

// localCacheEntry holds a cached value with expiration time.
type localCacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewLocalCache creates a new local cache with the given default TTL.
func NewLocalCache(ttl time.Duration) *LocalCache {
	lc := &LocalCache{ttl: ttl}
	// Start cleanup goroutine
	go lc.cleanup()
	return lc
}

// Get retrieves a value from the cache. Returns nil if not found or expired.
func (c *LocalCache) Get(key string) interface{} {
	val, ok := c.items.Load(key)
	if !ok {
		return nil
	}
	entry := val.(*localCacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.items.Delete(key)
		return nil
	}
	return entry.value
}

// Set stores a value in the cache with the default TTL.
func (c *LocalCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *LocalCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.items.Store(key, &localCacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	})
}

// Delete removes a value from the cache.
func (c *LocalCache) Delete(key string) {
	c.items.Delete(key)
}

// cleanup periodically removes expired entries.
func (c *LocalCache) cleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.items.Range(func(key, value interface{}) bool {
			entry := value.(*localCacheEntry)
			if now.After(entry.expiresAt) {
				c.items.Delete(key)
			}
			return true
		})
	}
}
