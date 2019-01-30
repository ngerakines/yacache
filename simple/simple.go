package simple

import (
	"time"

	"github.com/ngerakines/yacache"
)

// Key is a simple key implementation that is backed by the Golang string
// type.
type Key string

// Item is a simple item implementation.
type Item struct {
	value    interface{}
	err      error
	cached   time.Time
	duration time.Duration
}

// CacheableValue is a Cacheable structure for values (non-errors).
type CacheableValue struct {
	value    interface{}
	duration time.Duration
}

// CacheableError is a Cacheable structure for errors.
type CacheableError struct {
	err      error
	duration time.Duration
}

// NewErrorItem returns an Item structure for a value (non-error), ensuring
// it conforms to the yacache Item interface.
func NewItem(value interface{}, cached time.Time, duration time.Duration) yacache.Item {
	return Item{
		value:    value,
		err:      nil,
		cached:   cached,
		duration: duration,
	}
}

// NewErrorItem returns an Item structure for an error, ensuring it conforms
// to the yacache Item interface.
func NewErrorItem(err error, cached time.Time, duration time.Duration) yacache.Item {
	return Item{
		value:    nil,
		err:      err,
		cached:   cached,
		duration: duration,
	}
}

// NewCacheableValue returns a Cacheable structure for a value (non-error),
// ensuring it conforms to the yacache Cacheable interface.
func NewCacheableValue(value interface{}, duration time.Duration) yacache.Cacheable {
	return CacheableValue{
		value:    value,
		duration: duration,
	}
}

// NewCacheableValue returns a Cacheable structure for an error, ensuring it
// conforms to the yacache Cacheable interface.
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
