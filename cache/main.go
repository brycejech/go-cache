package cache

type Cache interface {
	Get(path []string) Cache
	Set(path []string, ttl int, data []byte) bool
	Read() []byte
	Delete(path []string) bool
}
