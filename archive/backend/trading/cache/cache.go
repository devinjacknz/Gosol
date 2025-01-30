package cache

import (
	"sync"
	"time"
)

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Size        int
	LoadAvg     float64
	LastUpdated time.Time
}

// Cache represents a thread-safe in-memory cache
type Cache struct {
	data       map[string]*cacheEntry
	mutex      sync.RWMutex
	stats      CacheStats
	statsMutex sync.RWMutex
}

type cacheEntry struct {
	value       interface{}
	expiration  time.Time
	lastAccess  time.Time
	accessCount int64
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
		c.recordMiss()
		return nil, false
	}

	now := time.Now()
	if now.After(entry.expiration) {
		go c.delete(key) // Async cleanup
		c.recordMiss()
		return nil, false
	}

	// Update access statistics
	entry.lastAccess = now
	entry.accessCount++
	c.recordHit()

	return entry.value, true
}

// GetStats returns current cache statistics
func (c *Cache) GetStats() CacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()
	return c.stats
}

func (c *Cache) recordHit() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	c.stats.Hits++
	c.stats.LastUpdated = time.Now()
}

func (c *Cache) recordMiss() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	c.stats.Misses++
	c.stats.LastUpdated = time.Now()
}

func (c *Cache) recordEviction() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	c.stats.Evictions++
	c.stats.LastUpdated = time.Now()
}

func (c *Cache) updateSize() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	c.stats.Size = len(c.data)
	c.stats.LastUpdated = time.Now()
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
	
	evictions := len(c.data)
	c.data = make(map[string]*cacheEntry)
	
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	c.stats.Evictions += int64(evictions)
	c.stats.Size = 0
	c.stats.LastUpdated = time.Now()
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
	
	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		c.recordEviction()
		c.updateSize()
	}
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

// SetPrice adds a price to the cache with LRU eviction
func (pc *PriceCache) SetPrice(token string, price float64, expiration time.Duration) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	// Check if we need to remove old entries
	if len(pc.data) >= pc.maxSize {
		// Find least recently used entry
		var lruKey string
		var lruTime time.Time
		first := true

		for key, entry := range pc.data {
			if first || entry.lastAccess.Before(lruTime) {
				lruKey = key
				lruTime = entry.lastAccess
				first = false
			}
		}

		if lruKey != "" {
			delete(pc.data, lruKey)
			pc.recordEviction()
		}
	}

	// Add new price
	now := time.Now()
	pc.data[token] = &cacheEntry{
		value:      price,
		expiration: now.Add(expiration),
		lastAccess: now,
	}
	pc.updateSize()
}

// GetPrice retrieves a price from the cache with access time update
func (pc *PriceCache) GetPrice(token string) (float64, bool) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	entry, exists := pc.data[token]
	if !exists {
		pc.recordMiss()
		return 0, false
	}

	now := time.Now()
	if now.After(entry.expiration) {
		delete(pc.data, token)
		pc.recordEviction()
		pc.recordMiss()
		return 0, false
	}

	// Update access time and count
	entry.lastAccess = now
	entry.accessCount++
	pc.recordHit()

	price, ok := entry.value.(float64)
	return price, ok
}

// GetCacheStats returns cache statistics for price cache
func (pc *PriceCache) GetCacheStats() CacheStats {
	pc.statsMutex.RLock()
	defer pc.statsMutex.RUnlock()

	stats := pc.stats
	stats.LoadAvg = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	return stats
}

// AnalysisCache represents a specialized cache for analysis results
type AnalysisCache struct {
	*Cache
	maxSize int
}

// NewAnalysisCache creates a new analysis cache
func NewAnalysisCache(maxSize int) *AnalysisCache {
	return &AnalysisCache{
		Cache:   New(),
		maxSize: maxSize,
	}
}

// SetAnalysis adds analysis results to the cache with LRU eviction
func (ac *AnalysisCache) SetAnalysis(key string, analysis interface{}, expiration time.Duration) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	// Check if we need to remove old entries
	if len(ac.data) >= ac.maxSize {
		// Find least recently used entry
		var lruKey string
		var lruTime time.Time
		first := true

		for key, entry := range ac.data {
			if first || entry.lastAccess.Before(lruTime) {
				lruKey = key
				lruTime = entry.lastAccess
				first = false
			}
		}

		if lruKey != "" {
			delete(ac.data, lruKey)
			ac.recordEviction()
		}
	}

	// Add new analysis
	now := time.Now()
	ac.data[key] = &cacheEntry{
		value:      analysis,
		expiration: now.Add(expiration),
		lastAccess: now,
	}
	ac.updateSize()
}

// GetAnalysis retrieves analysis results from the cache with access time update
func (ac *AnalysisCache) GetAnalysis(key string) (interface{}, bool) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	entry, exists := ac.data[key]
	if !exists {
		ac.recordMiss()
		return nil, false
	}

	now := time.Now()
	if now.After(entry.expiration) {
		delete(ac.data, key)
		ac.recordEviction()
		ac.recordMiss()
		return nil, false
	}

	// Update access time and count
	entry.lastAccess = now
	entry.accessCount++
	ac.recordHit()

	return entry.value, true
}

// GetAnalysisCacheStats returns cache statistics for analysis cache
func (ac *AnalysisCache) GetAnalysisCacheStats() CacheStats {
	ac.statsMutex.RLock()
	defer ac.statsMutex.RUnlock()

	stats := ac.stats
	stats.LoadAvg = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	return stats
}
