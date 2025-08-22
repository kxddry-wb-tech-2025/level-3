package cache

import (
	"container/list"
	c "context"
	"eventbooker/src/internal/storage"
	"sync"
	"time"
)

type cacheEntry struct {
	value   any
	time    time.Time
	lruElem *list.Element
}

// Cache uses TTL + LRU for cache invalidation.
// It is thread-safe.
// Note that there are no tests for this cache. This is because I've already comprehensively tested it in Level-0.
// I've used it in the previous project, so I'm confident in its correctness.
type Cache struct {
	mp       map[string]*cacheEntry
	mu       *sync.Mutex
	ttl      time.Duration
	stopChan chan struct{}
	limit    int
	lru      *list.List
}

// NewCache creates cache
func NewCache(ttl time.Duration, limit int) *Cache {
	cc := &Cache{
		mp:       make(map[string]*cacheEntry),
		mu:       new(sync.Mutex),
		ttl:      ttl,
		stopChan: make(chan struct{}),
		limit:    limit,
		lru:      list.New(),
	}

	go cc.removeExpired()
	return cc
}

// Set sets a value in the cache
func (c *Cache) Set(ctx c.Context, key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.mp[key]; ok {
		entry.value = value
		entry.time = time.Now()
		c.lru.MoveToFront(entry.lruElem)
		return nil
	}

	elem := c.lru.PushFront(key)
	c.mp[key] = &cacheEntry{
		value:   value,
		time:    time.Now(),
		lruElem: elem,
	}

	if c.lru.Len() > c.limit {
		c.removeLRU()
	}

	return nil
}

func (c *Cache) removeLRU() {
	back := c.lru.Back()
	if back == nil {
		return
	}
	orderID := back.Value.(string)
	c.remove(orderID)
}

func (c *Cache) remove(orderID string) {
	entry, ok := c.mp[orderID]
	if !ok {
		return
	}
	c.lru.Remove(entry.lruElem)
	delete(c.mp, orderID)
}

// Get gets a value from the cache
func (c *Cache) Get(ctx c.Context, key string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.mp[key]
	if !ok {
		return nil, storage.ErrNotFound
	}

	// cache invalidation
	if time.Since(entry.time) > c.ttl {
		c.remove(key)
		return nil, storage.ErrNotFound
	}

	c.lru.MoveToFront(entry.lruElem)

	return entry.value, nil
}

func (c *Cache) removeExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for id, entry := range c.mp {
				if now.Sub(entry.time) > c.ttl {
					c.remove(id)
				}
			}
			c.mu.Unlock()
		case <-c.stopChan:
			return
		}

	}
}

// Stop stops the cache
func (c *Cache) Stop() {
	close(c.stopChan)
}
