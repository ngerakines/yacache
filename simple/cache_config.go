package simple

import "github.com/ngerakines/yacache"

type CacheOption func(cache *Cache) error

// WithMaxSize configures the maximum number of elements that the cache
// will contain.
func WithMaxSize(size int) func(cache *Cache) error {
	return func(cache *Cache) error {
		cache.maxSize = size
		return nil
	}
}

// WithEvictionHandler configures the eviction callback function for the
// cache.
func WithEvictionHandler(callback yacache.EvictionCallback) func(cache *Cache) error {
	return func(cache *Cache) error {
		cache.evictionCallback = callback
		return nil
	}
}
