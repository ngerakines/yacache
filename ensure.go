package yacache

import "context"

// EnsureCacheGet retrieves an item from the cache, but panics if an error is
// returned.
func EnsureCacheGet(cache Cache, ctx context.Context, key Key, fetcher Fetcher) Item {
	item, err := cache.Get(ctx, key, fetcher)
	if err != nil {
		panic(err)
	}
	return item
}

// EnsureCachePut stores an item into the cache, but panics if an error is
// returned.
func EnsureCachePut(cache Cache, ctx context.Context, key Key, fetcher Fetcher) {
	err := cache.Put(ctx, key, fetcher)
	if err != nil {
		panic(err)
	}
}

// EnsureCacheContains returns the presense of a cached item, but panics if an
// error is returned.
func EnsureCacheContains(cache Cache, ctx context.Context, key Key) bool {
	contains, err := cache.Contains(ctx, key)
	if err != nil {
		panic(err)
	}
	return contains
}

// EnsureCacheDelete removes an item fromthe cache, but panics if an error is returned.
func EnsureCacheDelete(cache Cache, ctx context.Context, key Key) {
	err := cache.Delete(ctx, key)
	if err != nil {
		panic(err)
	}
}
