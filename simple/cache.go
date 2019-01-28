package simple

import (
	"context"
	"github.com/ngerakines/yacache"
	"sync"
	"time"
)

type Cache struct {
	values map[string]yacache.Item
	mu     sync.Mutex
}

func NewCache() yacache.Cache {
	return &Cache{
		values: make(map[string]yacache.Item),
	}
}

func (c *Cache) Get(ctx context.Context, key yacache.Key, fetcher yacache.Fetcher) (yacache.Item, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, hasItem := c.values[key.Value()]
	if hasItem {
		return item, nil
	}

	cacheable, err := fetcher(ctx, key)
	if err != nil {
		return nil, err
	}

	item = itemFromCacheable(cacheable)
	c.values[key.Value()] = item

	return item, nil
}

func (c *Cache) Put(ctx context.Context, key yacache.Key, fetcher yacache.Fetcher) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheable, err := fetcher(ctx, key)
	if err != nil {
		return err
	}

	item := itemFromCacheable(cacheable)
	c.values[key.Value()] = item

	return nil
}

func (c *Cache) Contains(ctx context.Context, key yacache.Key) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, hasItem := c.values[key.Value()]

	return hasItem, nil
}

func itemFromCacheable(item yacache.Cacheable) yacache.Item {
	if err := item.Error(); err != nil {
		return NewErrorItem(err, time.Now(), item.Duration())
	}
	return NewItem(item.Value(), time.Now(), item.Duration())
}
