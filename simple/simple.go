package simple

import (
	"github.com/ngerakines/yacache"
	"time"
)

type Key string

type Item struct {
	value    interface{}
	err      error
	cached   time.Time
	duration time.Duration
}

type CacheableValue struct {
	value    interface{}
	duration time.Duration
}

type CacheableError struct {
	err      error
	duration time.Duration
}

func NewItem(value interface{}, cached time.Time, duration time.Duration) yacache.Item {
	return Item{
		value:    value,
		err:      nil,
		cached:   cached,
		duration: duration,
	}
}

func NewErrorItem(err error, cached time.Time, duration time.Duration) yacache.Item {
	return Item{
		value:    nil,
		err:      err,
		cached:   cached,
		duration: duration,
	}
}

func NewCacheableValue(value interface{}, duration time.Duration) yacache.Cacheable {
	return CacheableValue{
		value:    value,
		duration: duration,
	}
}

func NewCacheableError(err error, duration time.Duration) yacache.Cacheable {
	return CacheableError{
		err:      err,
		duration: duration,
	}
}

func (i Item) Value() interface{} {
	return i.value
}

func (i Item) Error() error {
	return i.err
}

func (i Item) Cached() time.Time {
	return i.cached
}

func (i Item) Duration() time.Duration {
	return i.duration
}

func (i Item) Expired() bool {
	return time.Now().After(i.cached.Add(i.duration))
}

func (k Key) Value() string {
	return string(k)
}

func (c CacheableValue) Value() interface{} {
	return c.value
}

func (c CacheableValue) Error() error {
	return nil
}

func (c CacheableValue) Duration() time.Duration {
	return c.duration
}

func (c CacheableError) Value() interface{} {
	return nil
}

func (c CacheableError) Error() error {
	return c.err
}

func (c CacheableError) Duration() time.Duration {
	return c.duration
}
