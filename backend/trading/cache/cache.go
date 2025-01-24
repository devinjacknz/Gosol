package cache

import (
	"sync"
	"time"
)

// Cache represents a thread-safe in-memory cache
type Cache struct {
	data  map[string]*cacheEntry
	mutex sync.RWMutex
}

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// New creates a new cache instance
func New() *Cache {
	return &Cache{
		data: make(map[string]*cacheEntry),
	}
}

// Set adds a value to the cache with expiration
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &cacheEntry{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		go c.delete(key) // Async cleanup
		return nil, false
	}

	return entry.value, true
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// Clear removes all values from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]*cacheEntry)
}

// Cleanup removes expired entries from the cache
func (c *Cache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.expiration) {
			delete(c.data, key)
		}
	}
}

// StartCleanupTask starts periodic cleanup of expired entries
func (c *Cache) StartCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.Cleanup()
		}
	}()
}

// delete is an internal method for async cleanup
func (c *Cache) delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// PriceCache represents a specialized cache for price data
type PriceCache struct {
	*Cache
	maxSize int
}

// NewPriceCache creates a new price cache with maximum size
func NewPriceCache(maxSize int) *PriceCache {
	return &PriceCache{
		Cache:   New(),
		maxSize: maxSize,
	}
}

// SetPrice adds a price to the cache with size management
func (pc *PriceCache) SetPrice(token string, price float64, expiration time.Duration) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	// Check if we need to remove old entries
	if len(pc.data) >= pc.maxSize {
		// Remove oldest entry
		var oldestKey string
		var oldestTime time.Time
		first := true

		for key, entry := range pc.data {
			if first || entry.expiration.Before(oldestTime) {
				oldestKey = key
				oldestTime = entry.expiration
				first = false
			}
		}

		if oldestKey != "" {
			delete(pc.data, oldestKey)
		}
	}

	// Add new price
	pc.data[token] = &cacheEntry{
		value:      price,
		expiration: time.Now().Add(expiration),
	}
}

// GetPrice retrieves a price from the cache
func (pc *PriceCache) GetPrice(token string) (float64, bool) {
	value, exists := pc.Get(token)
	if !exists {
		return 0, false
	}

	price, ok := value.(float64)
	return price, ok
}

// AnalysisCache represents a specialized cache for analysis results
type AnalysisCache struct {
	*Cache
}

// NewAnalysisCache creates a new analysis cache
func NewAnalysisCache() *AnalysisCache {
	return &AnalysisCache{
		Cache: New(),
	}
}

// SetAnalysis adds analysis results to the cache
func (ac *AnalysisCache) SetAnalysis(key string, analysis interface{}, expiration time.Duration) {
	ac.Set(key, analysis, expiration)
}

// GetAnalysis retrieves analysis results from the cache
func (ac *AnalysisCache) GetAnalysis(key string) (interface{}, bool) {
	return ac.Get(key)
}
