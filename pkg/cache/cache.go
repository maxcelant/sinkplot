package cache

import (
	"iter"
	"sync"
)

type List[T any] []T

type Cache[T any] interface {
	Set(string, T)
	Get(string) (T, bool)
	Del(string)
	Items() iter.Seq2[string, T]
}

// Thread-safe cache to use for our registry so that we can be assured that concurrent writes don't
// end in unexpected ways
type cache[T any] struct {
	mu    sync.RWMutex
	table map[string]T
}

func New[T any]() Cache[T] {
	return &cache[T]{
		table: make(map[string]T),
	}
}

func (c *cache[T]) Set(k string, v T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.table[k] = v
}

func (c *cache[T]) Get(k string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.table[k]
	return v, ok
}

func (c *cache[T]) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.table, k)
}

func (c *cache[T]) Items() iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		for k, v := range c.table {
			if !yield(k, v) {
				return
			}
		}
	}
}
