package memo

import "sync"

type Cache[K comparable, V any] struct {
	mu     sync.Mutex
	values map[K]V
}

func (c *Cache[K, V]) Get(key K, load func() (V, error)) (V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.values[key]; ok {
		return v, nil
	}
	v, err := load()
	if err != nil {
		var zero V
		return zero, err
	}
	if c.values == nil {
		c.values = make(map[K]V)
	}
	c.values[key] = v
	return v, nil
}
