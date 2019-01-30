package redis

import (
	"context"
	"flag"
	"fmt"
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
	ctx := context.Background()

	key := simple.Key("foo")
	key2 := simple.Key("bar")
	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return simple.NewCacheableValue("value", 1*time.Hour), nil
	}

	redisClient := redisClient(t, 1)

	c := NewCache(redisClient, WithMaxSize(5))

	ok, err := c.Contains(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("key '%s' should not be in the cache", key)
	}

	err = c.Put(ctx, key, fetcher)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = c.Contains(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("key '%s' should be in the cache but is not", key)
	}

	ok, err = c.Contains(ctx, key2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("key '%s' should not be in the cache", key)
	}

	item, err := c.Get(ctx, key2, fetcher)
	if err != nil {
		t.Fatal(err)
	}
	if item == nil {
		t.Fatalf("key '%s' should be in the cache but is not", key)
	}
	if fmt.Sprintf("%s", item.Value()) != "value" {
		t.Fatalf("key '%s' returned unexpected item: %s", key, item.Value())
	}

	item, err = c.Get(ctx, key2, fetcher)
	if err != nil {
		t.Fatal(err)
	}
	if item == nil {
		t.Fatalf("key '%s' should be in the cache but is not", key)
	}
	if fmt.Sprintf("%s", item.Value()) != "value" {
		t.Fatalf("key '%s' returned unexpected item: %s", key, item.Value())
	}

	if err = c.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}

	ok, err = c.Contains(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("key '%s' should not be in the cache", key)
	}
}

func TestCacheMaxSize(t *testing.T) {
	ctx := context.Background()

	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return simple.NewCacheableValue("value", 1*time.Hour), nil
	}

	redisClient := redisClient(t, 1)

	c := NewCache(
		redisClient,
		WithMaxSize(5))

	for i := 10; i >= 6; i-- {
		_, err := c.Get(ctx, simple.Key(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 10; i++ {
		_, err := c.Get(ctx, simple.Key(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 5; i++ {
		if exists, err := c.Contains(ctx, simple.Key(fmt.Sprintf("%d", i))); err != nil || exists {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatalf("unexpected key: %d", i)
		}
	}
	for i := 6; i <= 10; i++ {
		if exists, err := c.Contains(ctx, simple.Key(fmt.Sprintf("%d", i))); err != nil || !exists {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatalf("expected key: %d", i)
		}
	}
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
