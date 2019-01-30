package redis

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/ngerakines/yacache"
	"github.com/ngerakines/yacache/simple"
	"sync"
	"time"
)

type Cache struct {
	Redis     *redis.Client
	Marshal   yacache.Marshaller
	Unmarshal yacache.Unmarshaller

	maxSize int64
	mu      sync.Mutex
}

const (
	hitsMetaKey       = "hits"
	valueAttribute    = "v"
	createdAttribute  = "c"
	durationAttribute = "d"
)

func NewCache(redisClient *redis.Client, options ...CacheOption) yacache.Cache {
	cache := &Cache{
		Redis:   redisClient,
		maxSize: -1,
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

	get, err := c.Redis.HGetAll(kv).Result()
	if getValue, ok := get[valueAttribute]; ok {
		if c.maxSize > -1 {
			_, err = c.Redis.ZAddXX(hitsMetaKey, redis.Z{Score: float64(now.UnixNano()), Member: kv}).Result()
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

	cacheable, err := fetcher(ctx, key)
	if err != nil {
		return nil, err
	}

	item := simple.ItemFromCacheable(cacheable)
	_, err = c.Redis.Pipelined(func(pipeliner redis.Pipeliner) error {
		pipeliner.HMSet(kv, map[string]interface{}{
			valueAttribute:    item.Value(),
			createdAttribute:  item.Cached().Format(time.RFC822Z),
			durationAttribute: item.Duration().String(),
		})
		pipeliner.Expire(kv, item.Duration())
		if c.maxSize > -1 {
			pipeliner.ZAdd("hits", redis.Z{Score: float64(now.UnixNano()), Member: kv})
		}
		return nil
	})
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

	kv := key.Value()
	now := time.Now()

	cacheable, err := fetcher(ctx, key)
	if err != nil {
		return err
	}

	item := simple.ItemFromCacheable(cacheable)
	_, err = c.Redis.Pipelined(func(pipeliner redis.Pipeliner) error {
		pipeliner.HMSet(kv, map[string]interface{}{
			valueAttribute:    item.Value(),
			createdAttribute:  item.Cached().Format(time.RFC822Z),
			durationAttribute: item.Duration().String(),
		})
		pipeliner.Expire(kv, item.Duration())
		if c.maxSize > -1 {
			pipeliner.ZAdd("hits", redis.Z{Score: float64(now.UnixNano()), Member: kv})
		}
		return nil
	})
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

	return c.Redis.HExists(key.Value(), valueAttribute).Result()
}

func (c *Cache) Delete(ctx context.Context, key yacache.Key) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	kv := key.Value()

	_, err := c.Redis.Pipelined(func(pipeliner redis.Pipeliner) error {
		pipeliner.Del(kv)
		pipeliner.ZRem(hitsMetaKey, kv)
		return nil
	})
	return err
}

func (c *Cache) clearExtra() error {
	if c.maxSize > 0 {
		count, err := c.Redis.ZCount(hitsMetaKey, "-inf", "+inf").Result()
		if err != nil {
			return err
		}
		if count > c.maxSize {
			scores, err := c.Redis.ZRevRange(hitsMetaKey, c.maxSize, -1).Result()
			if err != nil {
				return err
			}
			_, err = c.Redis.Pipelined(func(pipeliner redis.Pipeliner) error {
				pipeliner.Del(scores...)
				for _, score := range scores {
					pipeliner.ZRem(hitsMetaKey, score)
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
