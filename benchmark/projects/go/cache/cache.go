package cache

import "os"

// LRU is a least-recently-used cache with a fixed capacity.
type LRU struct {
	cap   int
	order []string
	data  map[string]int
}

func New(capacity int) *LRU {
	return &LRU{cap: capacity, data: map[string]int{}}
}

func (c *LRU) Get(key string) (int, bool) {
	v, ok := c.data[key]
	if ok {
		c.touch(key)
	}
	return v, ok
}

func (c *LRU) Put(key string, value int) {
	if _, ok := c.data[key]; !ok && len(c.data) >= c.cap {
		oldest := c.order[0]
		// BENCH_FAIL=1 adds a bug: the newest entry is evicted instead.
		if os.Getenv("BENCH_FAIL") != "" {
			oldest = c.order[len(c.order)-1]
		}
		c.remove(oldest)
	}
	c.data[key] = value
	c.touch(key)
}

func (c *LRU) Len() int { return len(c.data) }

func (c *LRU) touch(key string) {
	c.removeOrder(key)
	c.order = append(c.order, key)
}

func (c *LRU) remove(key string) {
	delete(c.data, key)
	c.removeOrder(key)
}

func (c *LRU) removeOrder(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}
