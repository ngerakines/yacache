package simple

import (
	"context"
	"sync"
	"time"

	"github.com/ngerakines/yacache"
)

type Cache struct {
	keys   []string
	values map[interface{}]yacache.Item

	maxSize          int
	evictionCallback yacache.EvictionCallback

	mu sync.Mutex
}

func NewCache(options ...CacheOption) yacache.Cache {
	cache := &Cache{
		keys:             make([]string, 0),
		values:           make(map[interface{}]yacache.Item),
		maxSize:          -1,
		evictionCallback: nil,
	}

	for _, option := range options {
		option(cache)
	}

	return cache
}

func (c *Cache) Get(ctx context.Context, key yacache.Key, fetcher yacache.Fetcher) (yacache.Item, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	kv := key.Value()

	item, hasItem := c.values[kv]
	if hasItem {
		c.keys = remove(c.keys, kv)
		c.keys = append(c.keys, kv)
		return item, nil
	}

	cacheable, err := fetcher(ctx, key)
	if err != nil {
		return nil, err
	}

	item = itemFromCacheable(cacheable)
	c.keys = append(c.keys, kv)
	c.values[kv] = item

	if c.maxSize > 0 && len(c.keys) > c.maxSize {
		c.pop()
	}

	return item, nil
}

func (c *Cache) Put(ctx context.Context, key yacache.Key, fetcher yacache.Fetcher) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	kv := key.Value()

	cacheable, err := fetcher(ctx, key)
	if err != nil {
		c.keys = remove(c.keys, kv)
		c.keys = append(c.keys, kv)
		return err
	}

	item := itemFromCacheable(cacheable)
	c.keys = append(c.keys, kv)
	c.values[kv] = item

	return nil
}

func (c *Cache) Contains(ctx context.Context, key yacache.Key) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, hasItem := c.values[key.Value()]

	return hasItem, nil
}

func (c *Cache) Delete(ctx context.Context, key yacache.Key) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	kv := key.Value()
	c.keys = remove(c.keys, kv)
	delete(c.values, kv)

	return nil
}

func (c *Cache) pop() {
	var key string
	key, c.keys = c.keys[0], c.keys[1:]

	item, hasItem := c.values[key]
	if !hasItem {
		return
	}

	delete(c.values, key)
	if c.evictionCallback != nil {
		c.evictionCallback(Key(key), item)
	}
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func itemFromCacheable(item yacache.Cacheable) yacache.Item {
	if err := item.Error(); err != nil {
		return NewErrorItem(err, time.Now(), item.Duration())
	}
	return NewItem(item.Value(), time.Now(), item.Duration())
}
