package cache

import (
    "context"
    "sync"
    "time"
)

// CacheKey is used as a type for cache keys.
type CacheKey string

// CacheEntry represents a single item in the cache with its value and expiration time.
type CacheEntry struct {
    Value      interface{}
    Expiration time.Time
}

// Config holds configuration parameters for initializing a Cache.
type Config struct {
    MaxSize       int           // Maximum number of items the cache can hold.
    TTL           time.Duration // Time to live for each cache item.
    CleanupPeriod time.Duration // How often to clean up expired items.
}

// CacheStats provides statistics on cache operations.
type CacheStats struct {
    Size    int   // Current number of items in the cache.
    Hits    int64 // Number of successful cache retrievals.
    Misses  int64 // Number of failed cache retrievals.
    Evicted int64 // Number of items removed from the cache due to eviction.
}

// Cache represents the cache structure, embedding sync.RWMutex for thread safety.
type Cache struct {
    sync.RWMutex
    items   map[CacheKey]CacheEntry
    maxSize int
    ttl     time.Duration
    ctx     context.Context
    cancel  context.CancelFunc
}

// New initializes and returns a new Cache instance with the given configuration.
func New(config Config) *Cache {
    ctx, cancel := context.WithCancel(context.Background())
    return &Cache{
        RWMutex: sync.RWMutex{},
        items:   make(map[CacheKey]CacheEntry),
        maxSize: config.MaxSize,
        ttl:     config.TTL,
        ctx:     ctx,
        cancel:  cancel,
    }
}
