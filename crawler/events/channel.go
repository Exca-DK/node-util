package events

import (
	"sync/atomic"
	"time"
)

type wrappedChan[T any] struct {
	ch    chan<- T
	tasks atomic.Int32
	done  atomic.Bool
}

func (c *wrappedChan[T]) Send(data T) {
	if c.done.Load() {
		return
	}
	c.tasks.Add(1)
	defer c.tasks.Add(-1)
	c.ch <- data
}

func (c *wrappedChan[T]) Close() {
	if !c.done.CompareAndSwap(false, true) {
		return
	}

	time.Sleep(1 * time.Second)
	if c.tasks.Load() > 0 {
		for {
			if c.tasks.Load() == 0 {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}
	close(c.ch)
}
