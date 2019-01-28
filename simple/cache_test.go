package simple

import (
	"context"
	"fmt"
	"github.com/ngerakines/yacache"
	"testing"
	"time"
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
