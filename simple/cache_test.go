package simple

import (
	"context"
	"fmt"
	"github.com/ngerakines/yacache/cachetest"
	"strconv"
	"testing"
	"time"

	"github.com/ngerakines/yacache"
)

func ExampleNewCache_Get() {
	ctx := context.Background()
	c := NewCache()
	key := Key("foo")
	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return NewCacheableValue("bar", 1*time.Hour), nil
	}
	if item, err := c.Get(ctx, key, fetcher); err == nil {
		fmt.Println(item.Value())
	}
	// Output: bar
}

func ExampleNewCache_Put() {
	ctx := context.Background()
	c := NewCache()
	key := Key("foo")
	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return NewCacheableValue("bar", 1*time.Hour), nil
	}
	fmt.Println(c.Put(ctx, key, fetcher))
	// Output: <nil>
}

func TestCache(t *testing.T) {
	c := NewCache()
	key := Key("foo")
	key2 := Key("bar")
	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return NewCacheableValue("value", 1*time.Hour), nil
	}
	cachetest.Standard(t, c, key, key2, fetcher)
}

func TestCacheMaxSize(t *testing.T) {
	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return NewCacheableValue("value", 1*time.Hour), nil
	}

	evictions := []string{}
	evictionCB := func(key yacache.Key, item yacache.Item) {
		evictions = append(evictions, key.Value())
	}

	c := NewCache(
		WithMaxSize(5),
		WithEvictionHandler(evictionCB),
	)

	cachetest.MaxSize(t, c, fetcher, func(s string) yacache.Key {
		return Key(s)
	})
	if len(evictions) != 10 {
		t.Fatalf("expected 10 evictions but there was %d", len(evictions))
	}
	for i, e := range []int{10, 9, 8, 7, 6, 1, 2, 3, 4, 5} {
		if strconv.Itoa(e) != evictions[i] {
			t.Fatalf("expected eviction at %d to be %s but got %d", i, evictions[i], e)
		}
	}
}
