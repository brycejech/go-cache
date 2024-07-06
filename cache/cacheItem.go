package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/brycejech/go-cache/util"
)

func newCacheItem() Cache {
	return &cacheItem{
		items: map[string]Cache{},
	}
}

type cacheItem struct {
	mu    sync.RWMutex
	ttl   int
	data  []byte
	items map[string]Cache
}

func (c *cacheItem) Set(path []string, ttl int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(path) == 0 {
		c.ttl = ttl
		c.data = data
		c.items = map[string]Cache{}
		return nil
	}

	if len(path) == 1 {
		time.AfterFunc(time.Duration(ttl)*time.Millisecond, func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			delete(c.items, path[0])
		})

		newItem := newCacheItem()
		newItem.Set([]string{}, ttl, data)

		c.items[path[0]] = newItem

		return nil
	}

	newItem := newCacheItem()
	c.items[path[0]] = newItem

	return newItem.Set(path[1:], ttl, data)
}

func (c *cacheItem) Get(path []string) Cache {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(path) == 0 {
		return c
	}

	item := c.items[path[0]]
	if len(path) == 1 || item == nil {
		return item
	}

	return item.Get(path[1:])
}

func (c *cacheItem) Read() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data
}

func (c *cacheItem) Delete(path []string) error {
	if len(path) == 0 {
		return errors.New(ERR_DELETE_EMPTY_PATH)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(path) == 1 {
		delete(c.items, path[0])
		return nil
	}

	item := c.Get(path)
	if item == nil {
		return nil
	}

	return item.Delete(path[1:])
}

func (c *cacheItem) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	size := len(c.data)

	for _, cache := range c.items {
		if cache != nil {
			size += cache.Size()
		}
	}

	return size
}

func (c *cacheItem) Visualize() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := map[string]any{}
	tree := map[string]any{
		"size":  util.ByteSizeToStr(int64(c.Size())),
		"items": items,
	}

	for k, cache := range c.items {
		if cache == nil {
			items[k] = map[string]any{
				"size":  util.ByteSizeToStr(0),
				"items": map[string]any{},
			}
		}

		items[k] = cache.Visualize()
	}

	return tree
}
