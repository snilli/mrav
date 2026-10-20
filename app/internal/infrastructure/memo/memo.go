// Package memo keeps what a function gave for a key, for the data that the viewer only reads and that never changes while it
// runs, so that the pages that ask twice a second ask the database once.
package memo

import "sync"

// Cache is the values that Get has loaded, by key. The zero value is ready to use. What it holds is shared between all that
// ask for it, so they must not write to it.
type Cache[K comparable, V any] struct {
	mu     sync.Mutex
	values map[K]V
}

// Get is the value for the key: the one that load gave the first time, or what load gives now when there is none yet. A
// failure is not kept, so the next call tries again. Callers at once wait for one load, and do not each make one.
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
