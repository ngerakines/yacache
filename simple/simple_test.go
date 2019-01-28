package simple

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

func ExampleKey() {
	fmt.Println(Key("example"))
	// Output: example
}

func ExampleNewItem() {
	now := time.Now()
	i := NewItem("example", now, 1*time.Hour)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: example
	// <nil>
	// 1h0m0s
}

func ExampleNewErrorItem() {
	now := time.Now()
	i := NewErrorItem(errors.New("failure"), now, 10*time.Minute)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: <nil>
	// failure
	// 10m0s
}

func ExampleNewCacheableValue() {
	i := NewCacheableValue("value", 1*time.Hour)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: value
	// <nil>
	// 1h0m0s
}

func ExampleNewCacheableError() {
	i := NewCacheableError(errors.New("failure"), 10*time.Minute)
	fmt.Println(i.Value())
	fmt.Println(i.Error())
	fmt.Println(i.Duration())
	// Output: <nil>
	// failure
	// 10m0s
}
