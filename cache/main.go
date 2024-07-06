package cache

type Cache interface {
	Get(path []string) Cache
	Set(path []string, ttl int, data []byte) error
	Read() []byte
	Delete(path []string) error
	Size() int
	Visualize() map[string]any
}

const ERR_DELETE_EMPTY_PATH = "must provide key to delete"
const ERR_SET_EMPTY_PATH = "must provide key to set"
