package yacache

import (
	"context"
	"time"
)

// Cache stores data for periods of time.
type Cache interface {
	Get(ctx context.Context, key Key, fetcher Fetcher) (Item, error)
	Put(ctx context.Context, key Key, fetcher Fetcher) error
	Contains(ctx context.Context, key Key) (bool, error)
	Delete(ctx context.Context, key Key) error
}

// Item is a piece of data that has been cached.
type Item interface {
	// Value returns the data that has been cached.
	Value() interface{}

	// Error returns an error if an error was cached.
	Error() error

	// Cached returns the time that the data was cached.
	Cached() time.Time

	// Duration returns the amount of time that the data was intended on being cached.
	Duration() time.Duration

	// Expired returns true if the cache time and duration are before now.
	Expired() bool
}

// Key is a way to identify an Item in the Cache.
type Key interface {
	Value() string
}

// Cacheable is a piece of data that will be cached.
type Cacheable interface {
	// Value returns the value to be cached.
	Value() interface{}

	// Error returns the error to be cached.
	Error() error

	// Duration returns the amount of time to keep the item in the cache.
	Duration() time.Duration
}

type Fetcher func(ctx context.Context, key Key) (Cacheable, error)

type EvictionCallback func(key Key, item Item)
