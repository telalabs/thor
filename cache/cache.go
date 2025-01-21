package cache

import (
    "context"
    "sync"
    "sync/atomic"
    "time"
)

// CacheStats holds statistics about the cache operations.
type CacheStats struct {
    Size    int
    Hits    int64
    Misses  int64
    Evicted int64
}

// CacheEntry represents an item in the cache with its value and expiration time.
type CacheEntry struct {
    Value      interface{}
    Expiration time.Time
}

// CacheKey is a type alias for string to represent cache keys.
type CacheKey string

// Config holds configuration parameters for the cache.
type Config struct {
    MaxSize       int
    TTL           time.Duration
    CleanupPeriod time.Duration
}

// Cache is the main structure that holds all cache data and methods.
type Cache struct {
    items    map[CacheKey]CacheEntry
    maxSize  int
    ttl      time.Duration
    ctx      context.Context
    cancel   context.CancelFunc
    mu       sync.RWMutex
}

var (
    hits    int64
    misses  int64
    evicted int64
)

// New initializes a new Cache with the given configuration.
func New(config Config) *Cache {
    ctx, cancel := context.WithCancel(context.Background())
    c := &Cache{
        items:   make(map[CacheKey]CacheEntry),
        maxSize: config.MaxSize,
        ttl:     config.TTL,
        ctx:     ctx,
        cancel:  cancel,
    }

    go c.cleanup(config.CleanupPeriod)
    return c
}

// Set adds an item to the cache. If the cache is full, it evicts the oldest item.
func (c *Cache) Set(key CacheKey, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if len(c.items) >= c.maxSize {
        c.evictOldest()
    }

    c.items[key] = CacheEntry{
        Value:      value,
        Expiration: time.Now().Add(c.ttl),
    }
}

// Get retrieves an item from the cache. It returns the value and a boolean indicating if the key was found.
func (c *Cache) Get(key CacheKey) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, exists := c.items[key]
    if !exists || time.Now().After(entry.Expiration) {
        atomic.AddInt64(&misses, 1)
        return nil, false
    }

    atomic.AddInt64(&hits, 1)
    return entry.Value, true
}

// Delete removes an item from the cache.
func (c *Cache) Delete(key CacheKey) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, key)
}

// Clear empties the cache.
func (c *Cache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items = make(map[CacheKey]CacheEntry)
}

// GetStats returns statistics on cache performance.
func (c *Cache) GetStats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()

    return CacheStats{
        Size:    len(c.items),
        Hits:    atomic.LoadInt64(&hits),
        Misses:  atomic.LoadInt64(&misses),
        Evicted: atomic.LoadInt64(&evicted),
    }
}

// cleanup runs periodically to remove expired items from the cache.
func (c *Cache) cleanup(period time.Duration) {
    ticker := time.NewTicker(period)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            c.mu.Lock()
            now := time.Now()
            for key, entry := range c.items {
                if now.After(entry.Expiration) {
                    delete(c.items, key)
                    atomic.AddInt64(&evicted, 1)
                }
            }
            c.mu.Unlock()
        case <-c.ctx.Done():
            return
        }
    }
}

// evictOldest removes the oldest item from the cache.
func (c *Cache) evictOldest() {
    var oldestKey CacheKey
    var oldestTime time.Time

    for key, entry := range c.items {
        if oldestTime.IsZero() || entry.Expiration.Before(oldestTime) {
            oldestKey, oldestTime = key, entry.Expiration
        }
    }

    if !oldestTime.IsZero() {
        delete(c.items, oldestKey)
        atomic.AddInt64(&evicted, 1)
    }
}

// Close cancels the context to stop the cleanup goroutine.
func (c *Cache) Close() {
    c.cancel()
}
