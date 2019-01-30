package redis

import (
	"context"
	"flag"
	"fmt"
	"github.com/ngerakines/yacache/cachetest"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/ngerakines/yacache"
	"github.com/ngerakines/yacache/simple"
)

var (
	redisHost  string
	redisDebug bool
)

type testHelper interface {
	Helper()
	Fatal(args ...interface{})
}

func init() {
	flag.StringVar(&redisHost, "redis.host", "localhost:6379", "The redis host to test against.")
	flag.BoolVar(&redisDebug, "redis.debug", false, "Do debug stuff with the redis client.")
}

func TestCache(t *testing.T) {
	redisClient := redisClient(t, 1)

	c := NewCache(redisClient, WithMaxSize(5))

	key := simple.Key("foo")
	key2 := simple.Key("bar")

	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return simple.NewCacheableValue("value", 1*time.Hour), nil
	}

	cachetest.Standard(t, c, key, key2, fetcher)
}

func TestCacheWithPrefix(t *testing.T) {
	redisClient := redisClient(t, 1)

	c := NewCache(
		redisClient,
		WithPrefix("what"),
	)

	key := simple.Key("foo")
	key2 := simple.Key("bar")

	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return simple.NewCacheableValue("value", 1*time.Hour), nil
	}

	cachetest.Standard(t, c, key, key2, fetcher)
}

func TestCacheMaxSize(t *testing.T) {
	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return simple.NewCacheableValue("value", 1*time.Hour), nil
	}

	redisClient := redisClient(t, 1)

	c := NewCache(
		redisClient,
		WithMaxSize(5),
		WithPrefix("TestCacheMaxSize"))

	cachetest.MaxSize(t, c, fetcher, func(s string) yacache.Key {
		return simple.Key(s)
	})
}

func BenchmarkCacheGet(b *testing.B) {
	ctx := context.Background()

	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return simple.NewCacheableValue("value", 1*time.Hour), nil
	}

	redisClient := redisClient(b, 2)
	c := NewCache(redisClient)

	for i := 0; i < b.N; i++ {
		_, err := c.Get(ctx, simple.Key(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheGet_maxSize(b *testing.B) {
	ctx := context.Background()

	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return simple.NewCacheableValue("value", 1*time.Hour), nil
	}

	redisClient := redisClient(b, 2)

	c := NewCache(
		redisClient,
		WithMaxSize(50),
	)
	for i := 0; i < b.N; i++ {
		_, err := c.Get(ctx, simple.Key(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func redisClient(t testHelper, db int) *redis.Client {
	t.Helper()
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   db,
	})
	if redisDebug {
		redisClient.WrapProcess(func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
			return func(cmd redis.Cmder) error {
				fmt.Printf("starting processing: <%s>\n", cmd)
				err := old(cmd)
				fmt.Printf("finished processing: <%s>\n", cmd)
				return err
			}
		})
	}

	_, err := redisClient.Ping().Result()
	if err != nil {
		t.Fatal(err)
	}

	_, err = redisClient.FlushAll().Result()
	if err != nil {
		t.Fatal(err)
	}
	return redisClient
}
