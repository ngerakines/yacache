package cachetest

import (
	"context"
	"fmt"
	"github.com/ngerakines/yacache"
	"testing"
)

func Standard(t *testing.T, c yacache.Cache, key, key2 yacache.Key, fetcher yacache.Fetcher) {
	t.Helper()

	ctx := context.Background()

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

func MaxSize(t *testing.T, c yacache.Cache, fetcher yacache.Fetcher, keyFactory SimpleKeyFactory) {
	ctx := context.Background()

	for i := 10; i >= 6; i-- {
		_, err := c.Get(ctx, keyFactory(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 10; i++ {
		_, err := c.Get(ctx, keyFactory(fmt.Sprintf("%d", i)), fetcher)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 5; i++ {
		if exists, err := c.Contains(ctx, keyFactory(fmt.Sprintf("%d", i))); err != nil || exists {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatalf("unexpected key: %d", i)
		}
	}
	for i := 6; i <= 10; i++ {
		if exists, err := c.Contains(ctx, keyFactory(fmt.Sprintf("%d", i))); err != nil || !exists {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatalf("expected key: %d", i)
		}
	}
}
