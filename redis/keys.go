package redis

type KeyTransform func(string) string

func DefaultKeyTransform(key string) string {
	return key
}
