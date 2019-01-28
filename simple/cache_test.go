package simple

import (
	"context"
	"fmt"
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
	ctx := context.Background()

	key := Key("foo")
	key2 := Key("bar")
	fetcher := func(ctx context.Context, fkey yacache.Key) (yacache.Cacheable, error) {
		return NewCacheableValue("value", 1*time.Hour), nil
	}

	c := NewCache()

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
	for i := 10; i >= 6; i-- {
		_, err := c.Get(ctx, Key(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 10; i++ {
		_, err := c.Get(ctx, Key(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 5; i++ {
		if exists, err := c.Contains(ctx, Key(fmt.Sprintf("%d", i))); err != nil || exists {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatalf("unexpected key: %d", i)
		}
	}
	for i := 6; i <= 10; i++ {
		if exists, err := c.Contains(ctx, Key(fmt.Sprintf("%d", i))); err != nil || !exists {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatalf("expected key: %d", i)
		}
	}
	if len(evictions) != 10 {
		t.Fatalf("expected 10 evictions but there was %d", len(evictions))
	}
	for i, e := range []int{10, 9, 8, 7, 6, 1, 2, 3, 4, 5} {
		if strconv.Itoa(e) != evictions[i] {
			t.Fatalf("expected eviction at %d to be %s but got %d", i, evictions[i], e)
		}
	}
}
