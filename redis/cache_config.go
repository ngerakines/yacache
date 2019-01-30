package redis

type CacheOption func(cache *Cache) error

func WithMaxSize(size int64) func(cache *Cache) error {
	return func(cache *Cache) error {
		cache.maxSize = size
		return nil
	}
}
