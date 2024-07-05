package cache

/*
	Tree-based, thread-safe cache implementation using sync.Map

	Outer cache use sync.Map

	Tried out multiple combinations and fastest impl for avg read/write time
	seems to be sync.Map for outer cache with map[string]Cache item values
	and manual RWMutex management

	May also be worth adding to cache struct to make it truly recursive
	rather than using cacheItem
*/

import (
	"sync"
	"time"
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

	item, ok := mapItem.(*cacheItem)
	if !ok {
		return nil
	}

	if len(path) == 1 {
		return item
	}

	return item.Get(path[1:])
}

func (c *cache) Set(path []string, ttl int, data []byte) bool {
	if len(path) == 0 {
		return false
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
			mapCache := NewMapCache()
			mapCache.Set([]string{}, ttl, data)
			c.items.Store(key, mapCache)

			// c.items.Store(key, newCacheItem([]string{}, ttl, data))

			return true
		}

		mapCache := NewMapCache()
		mapCache.Set(path[1:], ttl, data)
		c.items.Store(key, mapCache)

		// c.items.Store(key, newCacheItem(path[1:], ttl, data))
		return true
	}

	item, ok := existing.(*cacheItem)
	if !ok {
		return false
	}

	if len(path) == 1 {
		return item.Set([]string{}, ttl, data)
	}

	return item.Set(path[1:], ttl, data)
}

func (c *cache) Delete(path []string) bool {
	if len(path) == 0 {
		return true
	}

	if len(path) == 1 {
		c.items.Delete(path[0])
		return true
	}

	item := c.Get(path)
	if item == nil {
		return true
	}

	return item.Delete(path[1:])
}

func (c *cache) Read() []byte {
	return []byte{}
}

func newCacheItem(path []string, ttl int, data []byte) Cache {
	item := &cacheItem{
		items: &sync.Map{},
	}

	if len(path) == 0 {
		item.ttl = ttl
		item.data = data

		return item
	}

	item.Set(path, ttl, data)

	return item
}

type cacheItem struct {
	mu    sync.RWMutex
	ttl   int
	data  []byte
	items *sync.Map
}

func (c *cacheItem) Get(path []string) Cache {
	if len(path) < 1 {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key := path[0]
	mapItem, ok := c.items.Load(key)
	if !ok {
		return nil
	}

	item, ok := mapItem.(*cacheItem)
	if !ok {
		return nil
	}

	if len(path) == 1 {
		return item
	}

	return item.Get(path[1:])
}

func (c *cacheItem) Set(path []string, ttl int, data []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(path) == 0 {
		c.ttl = ttl
		c.data = data
		c.items = &sync.Map{}

		return true
	}

	key := path[0]
	if len(path) == 1 {
		c.items.Store(key, newCacheItem([]string{}, ttl, data))

		time.AfterFunc(time.Millisecond*time.Duration(ttl), func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			c.items.Delete(key)
		})

		return true
	}

	c.items.Store(key, newCacheItem(path[1:], ttl, data))
	return true
}

func (c *cacheItem) Delete(path []string) bool {
	if len(path) == 0 {
		return true
	}

	if len(path) == 1 {
		c.items.Delete(path[0])
		return true
	}

	item := c.Get(path)
	if item == nil {
		return true
	}

	return item.Delete(path[1:])
}

func (c *cacheItem) Read() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data
}
