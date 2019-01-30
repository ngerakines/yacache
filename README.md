# yet another cache

[![Build Status](https://travis-ci.org/ngerakines/yacache.png?branch=master)](https://travis-ci.org/ngerakines/yacache)
[![GoDoc](https://godoc.org/github.com/ngerakines/yacache?status.svg)](https://godoc.org/github.com/ngerakines/yacache)

Cache some things and stuff.

## Installation

Install:

```shell
go get -u github.com/ngerakines/yacache
```

Import:

```go
import "github.com/ngerakines/yacache"
```

## Quickstart

```go
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
	if item, err := c.Get(ctx, key, fetcher); err == nil {
		fmt.Println(item.Value())
	}
	// Output: bar
	// bar
}
```
