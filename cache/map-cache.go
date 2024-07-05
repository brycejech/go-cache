package cache

import (
	"sync"
	"time"
)

type mapCache struct {
	mu    sync.RWMutex
	ttl   int
	data  []byte
	items map[string]Cache
}

func NewMapCache() Cache {
	return &mapCache{
		items: map[string]Cache{},
	}
}

func (c *mapCache) Set(path []string, ttl int, data []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(path) == 0 {
		c.ttl = ttl
		c.data = data
		c.items = map[string]Cache{}
		return true
	}

	if len(path) == 1 {
		time.AfterFunc(time.Duration(ttl)*time.Millisecond, func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			delete(c.items, path[0])
		})

		newItem := NewMapCache()
		newItem.Set([]string{}, ttl, data)

		c.items[path[0]] = newItem
		return true
	}

	newItem := NewMapCache()
	c.items[path[0]] = newItem

	return newItem.Set(path[1:], ttl, data)
}

func (c *mapCache) Get(path []string) Cache {
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

func (c *mapCache) Read() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data
}

func (c *mapCache) Delete(path []string) bool {
	if len(path) == 0 {
		return true
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(path) == 1 {
		delete(c.items, path[0])
		return true
	}

	item := c.Get(path)
	if item == nil {
		return true
	}

	return item.Delete(path[1:])
}
