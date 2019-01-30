package redis

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/ngerakines/yacache"
	"github.com/ngerakines/yacache/simple"
)

type purgeBehavior int8

const (
	lru = 0
	lfa = 1
)

type Cache struct {
	redisClient   *redis.Client
	maxSize       int64
	prefix        string
	mu            sync.Mutex
	keyTransform  KeyTransform
	purgeBehavior purgeBehavior
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
		redisClient:   redisClient,
		maxSize:       -1,
		prefix:        "",
		keyTransform:  DefaultKeyTransform,
		purgeBehavior: lru,
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
		if c.maxSize > 0 && c.purgeBehavior == lfa {
			_, err = c.redisClient.ZAddXX(c.keyTransform(hitsMetaKey), redis.Z{Score: float64(now.UnixNano()), Member: c.keyTransform(kv)}).Result()
			if err != nil {
				return nil, err
			}
		}

		createdInt, err := strconv.ParseInt(get[createdAttribute], 10, 64)
		if err != nil {
			return nil, err
		}
		created := time.Unix(0, createdInt)

		if err != nil {
			return nil, err
		}
		dur, err := time.ParseDuration(get[durationAttribute])

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
		if c.maxSize > 0 && c.purgeBehavior == lru {
			pipeliner.SRem(c.keyTransform(hitsMetaKey), c.keyTransform(kv))
		} else if c.maxSize > 0 && c.purgeBehavior == lru {
			pipeliner.ZRem(c.keyTransform(hitsMetaKey), c.keyTransform(kv))
		}
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
			createdAttribute:  item.Cached().UnixNano(),
			durationAttribute: item.Duration().String(),
		})
		pipe.Expire(c.keyTransform(kv), item.Duration())
		if c.maxSize > 0 && c.purgeBehavior == lru {
			pipe.SAdd(c.keyTransform(hitsMetaKey), c.keyTransform(kv))
		} else if c.maxSize > 0 && c.purgeBehavior == lfa {
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
	var keys []string
	var count int64
	var err error
	if c.maxSize > 0 {
		switch c.purgeBehavior {
		case lru:
			keys, err = c.redisClient.Sort(c.keyTransform(hitsMetaKey), redis.Sort{
				By:     "*->c",
				Offset: float64(c.maxSize),
				Count:  -1,
			}).Result()
			if err != nil {
				return err
			}
		case lfa:
			count, err = c.redisClient.ZCount(c.keyTransform(hitsMetaKey), "-inf", "+inf").Result()
			if err != nil {
				return err
			}
			if count > c.maxSize {
				keys, err = c.redisClient.ZRevRange(c.keyTransform(hitsMetaKey), c.maxSize, -1).Result()
				if err != nil {
					return err
				}
			}
		}
	}
	if len(keys) > 0 {
		_, err = c.redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			pipeliner.Del(keys...)
			if c.maxSize > 0 && c.purgeBehavior == lru {
				for _, score := range keys {
					pipeliner.SRem(c.keyTransform(hitsMetaKey), score)
				}
			} else if c.maxSize > 0 && c.purgeBehavior == lfa {
				for _, score := range keys {
					pipeliner.ZRem(c.keyTransform(hitsMetaKey), score)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
