package cache

import (
	"sync"
	"time"
)

// Cache is a concurrent cache structure.
type Cache struct {
	data  map[string]CacheItem
	mutex sync.RWMutex
}

// CacheItem represents an item in the cache.
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// New creates a new concurrent cache.
func New() *Cache {
	return &Cache{
		data: make(map[string]CacheItem),
	}
}

// Set adds an item to the cache with a specified key, value, and expiration time.
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expirationTime := time.Now().Add(expiration)
	c.data[key] = CacheItem{
		Value:      value,
		Expiration: expirationTime,
	}
}

// Get retrieves the value associated with a key from the cache.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()

	item, exists := c.data[key]
	if !exists {
		c.mutex.RUnlock()
		return nil, false
	}

	if time.Now().After(item.Expiration) {
		// Unlock before locking again
		c.mutex.RUnlock()

		// Acquire write mutex.
		c.mutex.Lock()
		delete(c.data, key)
		c.mutex.Unlock()
		return nil, false
	}

	c.mutex.RUnlock()

	return item.Value, true
}
