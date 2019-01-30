package yacache

import "fmt"

func DefaultMarshaller(value interface{}) ([]byte, error) {
	switch a := value.(type) {
	case []byte:
		return a, nil
	default:
		return nil, fmt.Errorf("unexpected type: %T", a)
	}
}

func DefaultUnmarshaller(in []byte, out interface{}) error {
	switch a := out.(type) {
	case *[]byte:
		a = &in
		return nil
	default:
		return fmt.Errorf("unexpected type: %T", a)
	}
}
