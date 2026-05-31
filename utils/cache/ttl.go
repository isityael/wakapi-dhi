package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Item struct {
	Object     any
	Expiration int64
}

type Cache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		defaultExpiration: defaultExpiration,
		items:             make(map[string]Item),
	}
}

func (c *Cache) SetDefault(key string, value any) {
	c.Set(key, value, DefaultExpiration)
}

func (c *Cache) Set(key string, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Object:     value,
		Expiration: c.expirationFor(duration),
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	if item.expired(time.Now()) {
		c.Delete(key)
		return nil, false
	}

	return item.Object, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item)
}

func (c *Cache) Items() map[string]Item {
	now := time.Now()
	result := make(map[string]Item)

	c.mu.RLock()
	for key, item := range c.items {
		if !item.expired(now) {
			result[key] = item
		}
	}
	c.mu.RUnlock()

	return result
}

func (c *Cache) ItemCount() int {
	return len(c.Items())
}

func (c *Cache) IncrementInt(key string, delta int) (int, error) {
	result, err := c.increment(key, int64(delta))
	if err != nil {
		return 0, err
	}
	return int(result), nil
}

func (c *Cache) IncrementInt64(key string, delta int64) (int64, error) {
	return c.increment(key, delta)
}

func (c *Cache) increment(key string, delta int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok || item.expired(time.Now()) {
		delete(c.items, key)
		return 0, errors.New("item not found")
	}

	switch value := item.Object.(type) {
	case int:
		next := value + int(delta)
		item.Object = next
		c.items[key] = item
		return int64(next), nil
	case int64:
		next := value + delta
		item.Object = next
		c.items[key] = item
		return next, nil
	default:
		return 0, fmt.Errorf("item %q is not an integer", key)
	}
}

func (c *Cache) expirationFor(duration time.Duration) int64 {
	if duration == DefaultExpiration {
		duration = c.defaultExpiration
	}
	if duration == NoExpiration {
		return 0
	}
	return time.Now().Add(duration).UnixNano()
}

func (i Item) expired(now time.Time) bool {
	return i.Expiration > 0 && now.UnixNano() > i.Expiration
}
