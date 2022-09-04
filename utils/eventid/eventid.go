package eventid

import (
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type Cache struct {
	ids     map[hash.Event]bool
	mu      sync.RWMutex
	maxSize int
	epoch   idx.Epoch
}

func NewCache(maxSize int) *Cache {
	return &Cache{
		maxSize: maxSize,
	}
}

func (c *Cache) Reset(epoch idx.Epoch) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ids = make(map[hash.Event]bool)
	c.epoch = epoch
}

func (c *Cache) Has(id hash.Event) (has bool, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.ids == nil {
		return false, false
	}
	if c.epoch != id.Epoch() {
		return false, false
	}
	return c.ids[id], true
}

func (c *Cache) Add(id hash.Event) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ids == nil {
		return false
	}
	if c.epoch != id.Epoch() {
		return false
	}
	if len(c.ids) >= c.maxSize {
		c.ids = nil
		return false
	}
	c.ids[id] = true
	return true
}

func (c *Cache) Remove(id hash.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ids == nil {
		return
	}
	delete(c.ids, id)
}
