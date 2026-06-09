package cache

import (
	"container/list"
	"sync"
)

type lruEntry[K comparable, V any] struct {
	key   K
	value V
}

type LRU[K comparable, V any] struct {
	capacity int
	entries  map[K]*list.Element
	order    *list.List
	mu       sync.Mutex
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity < 1 {
		panic("cache capacity must be positive")
	}

	return &LRU[K, V]{
		capacity: capacity,
		entries:  make(map[K]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *LRU[K, V]) Add(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.entries[key]; ok {
		element.Value.(*lruEntry[K, V]).value = value
		c.order.MoveToFront(element)
		return
	}

	element := c.order.PushFront(&lruEntry[K, V]{key: key, value: value})
	c.entries[key] = element

	if c.order.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *LRU[K, V]) Contains(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.entries[key]
	return ok
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.entries[key]; ok {
		c.order.MoveToFront(element)
		return element.Value.(*lruEntry[K, V]).value, true
	}

	var zero V
	return zero, false
}

func (c *LRU[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.entries[key]; ok {
		c.order.Remove(element)
		delete(c.entries, key)
	}
}

func (c *LRU[K, V]) removeOldest() {
	element := c.order.Back()
	if element == nil {
		return
	}

	entry := element.Value.(*lruEntry[K, V])
	c.order.Remove(element)
	delete(c.entries, entry.key)
}
