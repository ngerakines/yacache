package simple

import "github.com/ngerakines/yacache"

type CacheOption func(cache *Cache) error

func WithMaxSize(size int) func(cache *Cache) error {
	return func(cache *Cache) error {
		cache.maxSize = size
		return nil
	}
}

func WithEvictionHandler(callback yacache.EvictionCallback) func(cache *Cache) error {
	return func(cache *Cache) error {
		cache.evictionCallback = callback
		return nil
	}
}
