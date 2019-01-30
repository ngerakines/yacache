package redis

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/ngerakines/yacache"
	"github.com/ngerakines/yacache/simple"
)

type Cache struct {
	redisClient  *redis.Client
	maxSize      int64
	prefix       string
	mu           sync.Mutex
	keyTransform KeyTransform
}

const (
	hitsMetaKey       = "yacache:keys"
	valueAttribute    = "v"
	createdAttribute  = "c"
	durationAttribute = "d"
)

// NewCache returns a new yacache.Cache that is backed by Redis.
func NewCache(redisClient *redis.Client, options ...CacheOption) yacache.Cache {
	cache := &Cache{
		redisClient:  redisClient,
		maxSize:      -1,
		prefix:       "",
		keyTransform: DefaultKeyTransform,
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
	now := time.Now()

	get, err := c.redisClient.HGetAll(c.keyTransform(kv)).Result()
	if getValue, ok := get[valueAttribute]; ok {
		if c.maxSize > -1 {
			_, err = c.redisClient.ZAddXX(c.keyTransform(hitsMetaKey), redis.Z{Score: float64(now.UnixNano()), Member: kv}).Result()
			if err != nil {
				return nil, err
			}
		}
		created, err := time.Parse(time.RFC822Z, get[createdAttribute])
		if err != nil {
			return nil, err
		}
		dur, err := time.ParseDuration(get[durationAttribute])
		if err != nil {
			return nil, err
		}
		return simple.NewItem(getValue, created, dur), nil
	}

	item, err := c.getAndSet(ctx, key, fetcher)
	if err != nil {
		return nil, err
	}

	if err = c.clearExtra(); err != nil {
		return nil, err
	}

	return item, nil
}

func (c *Cache) Put(ctx context.Context, key yacache.Key, fetcher yacache.Fetcher) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.getAndSet(ctx, key, fetcher)
	if err != nil {
		return err
	}

	if err = c.clearExtra(); err != nil {
		return err
	}

	return nil
}

func (c *Cache) Contains(ctx context.Context, key yacache.Key) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.redisClient.HExists(c.keyTransform(key.Value()), valueAttribute).Result()
}

func (c *Cache) Delete(ctx context.Context, key yacache.Key) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	kv := key.Value()

	_, err := c.redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
		pipeliner.Del(c.keyTransform(kv))
		pipeliner.ZRem(c.keyTransform(hitsMetaKey), c.keyTransform(kv))
		return nil
	})
	return err
}

func (c *Cache) getAndSet(ctx context.Context, key yacache.Key, fetcher yacache.Fetcher) (yacache.Item, error) {
	kv := key.Value()
	now := time.Now()

	cacheable, err := fetcher(ctx, key)
	if err != nil {
		return nil, err
	}

	item := simple.ItemFromCacheable(cacheable)
	_, err = c.redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.HMSet(c.keyTransform(kv), map[string]interface{}{
			valueAttribute:    item.Value(),
			createdAttribute:  item.Cached().Format(time.RFC822Z),
			durationAttribute: item.Duration().String(),
		})
		pipe.Expire(c.keyTransform(kv), item.Duration())
		if c.maxSize > -1 {
			pipe.ZAdd(c.keyTransform(hitsMetaKey), redis.Z{Score: float64(now.UnixNano()), Member: c.keyTransform(kv)})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Cache) clearExtra() error {
	if c.maxSize > 0 {
		count, err := c.redisClient.ZCount(c.keyTransform(hitsMetaKey), "-inf", "+inf").Result()
		if err != nil {
			return err
		}
		if count > c.maxSize {
			scores, err := c.redisClient.ZRevRange(c.keyTransform(hitsMetaKey), c.maxSize, -1).Result()
			if err != nil {
				return err
			}
			_, err = c.redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
				pipeliner.Del(scores...)
				for _, score := range scores {
					pipeliner.ZRem(c.keyTransform(hitsMetaKey), score)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
