package redis

import "fmt"

type CacheOption func(cache *Cache) error

func WithMaxSize(size int64) func(cache *Cache) error {
	return func(cache *Cache) error {
		cache.maxSize = size
		return nil
	}
}

func WithPrefix(prefix string) func(cache *Cache) error {
	return func(cache *Cache) error {
		cache.keyTransform = func(key string) string {
			return fmt.Sprintf("%s:%s", prefix, key)
		}
		return nil
	}
}
