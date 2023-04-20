package crawl

import (
	"sync"
)

type buffer[T any] struct {
	mu    sync.Mutex
	inner []T
}

// Enqueue inserts a value into the list
func (b *buffer[T]) Enqueue(data T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inner = append(b.inner, data)
}

// Enqueue inserts a value into the list
func (b *buffer[T]) Dequeue() (T, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var t T
	if len(b.inner) == 0 {
		return t, false
	}
	elem := b.inner[0]
	b.inner[0] = t
	b.inner = b.inner[1:]
	return elem, true
}
