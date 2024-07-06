package cache

/*
	Tree-based, thread-safe, in-memory cache implementation utilizing sync.Map

	To increase read/write perf, inner cache items are a map[string]*cacheItem
	that use a *sync.RWMutext for locking
*/

import (
	"errors"
	"sync"
	"time"

	"github.com/brycejech/go-cache/util"
)

func NewCache() Cache {
	return &cache{
		items: &sync.Map{},
	}
}

type cache struct {
	items *sync.Map
}

func (c *cache) Get(path []string) Cache {
	if len(path) < 1 {
		return nil
	}

	key := path[0]
	mapItem, ok := c.items.Load(key)
	if !ok {
		return nil
	}

	item, ok := mapItem.(Cache)
	if !ok {
		return nil
	}

	if len(path) == 1 {
		return item
	}

	return item.Get(path[1:])
}

func (c *cache) Set(path []string, ttl int, data []byte) error {
	if len(path) == 0 {
		return errors.New(ERR_SET_EMPTY_PATH)
	}

	key := path[0]

	if len(path) == 1 {
		time.AfterFunc(time.Millisecond*time.Duration(ttl), func() {
			c.items.Delete(key)
		})
	}

	existing, ok := c.items.Load(key)
	if !ok {
		if len(path) == 1 {
			mapCache := newCacheItem()
			mapCache.Set([]string{}, ttl, data)
			c.items.Store(key, mapCache)

			return nil
		}

		mapCache := newCacheItem()
		mapCache.Set(path[1:], ttl, data)
		c.items.Store(key, mapCache)
	}

	item, ok := existing.(Cache)
	if !ok {
		return errors.New("assertion error: cacheItem is not Cache")
	}

	if len(path) == 1 {
		return item.Set([]string{}, ttl, data)
	}

	return item.Set(path[1:], ttl, data)
}

func (c *cache) Delete(path []string) error {
	if len(path) == 0 {
		return errors.New(ERR_DELETE_EMPTY_PATH)
	}

	if len(path) == 1 {
		c.items.Delete(path[0])
		return nil
	}

	item := c.Get(path)
	if item == nil {
		return nil
	}

	return item.Delete(path[1:])
}

func (c *cache) Read() []byte {
	return []byte{}
}

func (c *cache) Size() int {
	size := 0

	c.items.Range(func(_ any, val any) bool {
		cache, ok := val.(Cache)
		if !ok || cache == nil {
			return true
		}

		size += cache.Size()
		return true
	})

	return size
}

func (c *cache) Visualize() map[string]any {
	items := map[string]any{}
	tree := map[string]any{
		"size":  util.ByteSizeToStr(int64(c.Size())),
		"items": items,
	}

	c.items.Range(func(key any, val any) bool {
		cache, ok := val.(Cache)
		if !ok || cache == nil {
			items[key.(string)] = map[string]any{
				"size":  util.ByteSizeToStr(0),
				"items": map[string]any{},
			}
			return true
		}
		items[key.(string)] = cache.Visualize()
		return true
	})

	return tree
}
