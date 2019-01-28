package yacache

import (
	"context"
	"fmt"
	"testing"
)

type failingCache struct {
}

func (failingCache) Get(ctx context.Context, key Key, fetcher Fetcher) (Item, error) {
	return nil, fmt.Errorf("not implemented")
}

func (failingCache) Put(ctx context.Context, key Key, fetcher Fetcher) error {
	return fmt.Errorf("not implemented")
}

func (failingCache) Contains(ctx context.Context, key Key) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (failingCache) Delete(ctx context.Context, key Key) error {
	return fmt.Errorf("not implemented")
}

func TestEnsureCacheGet(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	EnsureCacheGet(failingCache{}, context.Background(), nil, nil)
}

func TestEnsureCachePut(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	EnsureCachePut(failingCache{}, context.Background(), nil, nil)
}

func TestEnsureCacheContains(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	EnsureCacheContains(failingCache{}, context.Background(), nil)
}

func TestEnsureCacheDelete(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	EnsureCacheDelete(failingCache{}, context.Background(), nil)
}
