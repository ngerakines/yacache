package simple

import (
	"fmt"
	"github.com/pkg/errors"
	"time"
)

func ExampleKey() {
	fmt.Println(Key("example"))
	// Output: example
}

func ExampleItem() {
	now := time.Now()
	i := NewItem("example", now, 1 * time.Hour)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: example
	// <nil>
	// 1h0m0s
}

func ExampleItem_error() {
	now := time.Now()
	i := NewErrorItem(errors.New("failure"), now, 10 * time.Minute)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: <nil>
	// failure
	// 10m0s
}

func ExampleCacheableValue() {
	i := NewCacheableValue("value", 1 * time.Hour)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: value
	// <nil>
	// 1h0m0s
}

func ExampleCacheableValue_error() {
	i := NewCacheableError(errors.New("failure"), 10 * time.Minute)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: <nil>
	// failure
	// 10m0s
}