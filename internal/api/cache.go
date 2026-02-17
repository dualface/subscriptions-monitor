package api

import (
	"sync"
	"time"

	"github.com/user/subscriptions-monitor/internal/provider"
)

type Cache struct {
	mu        sync.RWMutex
	data      []provider.UsageSnapshot
	updatedAt time.Time
	ttl       time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		ttl: ttl,
	}
}

func (c *Cache) Get() ([]provider.UsageSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil || time.Since(c.updatedAt) > c.ttl {
		return nil, false
	}
	return c.data, true
}

func (c *Cache) Set(data []provider.UsageSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = data
	c.updatedAt = time.Now()
}
