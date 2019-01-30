package yacache

import (
	"bytes"
	"testing"
)

func TestDefaultMarshaller(t *testing.T) {
	out, err := DefaultMarshaller([]byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, []byte("foo")) {
		t.Fatal("bytes not the same")
	}

	_, err = DefaultMarshaller("bar")
	if err.Error() != "unexpected type: string" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
