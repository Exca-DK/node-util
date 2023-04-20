package events

import (
	"sync"
	"sync/atomic"

	"github.com/Exca-DK/node-util/crawler/utils"
)

type Feed[T any] struct {
	listeners map[utils.UUID]*wrappedChan[T]

	subs    atomic.Int32
	mu      sync.RWMutex
	stopped uint32
}

func (f *Feed[T]) init() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.listeners = make(map[utils.UUID]*wrappedChan[T])
}

func (f *Feed[T]) Subscribe(ch chan T) Subscription {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.stopped != 0 {
		return &emptySubscription{}
	}
	id := utils.NewId()
	wrap := &wrappedChan[T]{ch: ch}
	f.listeners[id] = wrap
	f.subs.Add(1)
	return &subscription{uuid: id, f: f.Unsubscribe}
}

func (f *Feed[T]) Unsubscribe(id utils.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sub, ok := f.listeners[id]
	if !ok {
		return
	}
	f.subs.Add(-1)
	delete(f.listeners, id)
	sub.Close()
}

func (f *Feed[T]) Send(data T) {
	if f.subs.Load() == 0 {
		return
	}
	go f.send(data)
}

func (f *Feed[T]) send(data T) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for i := range f.listeners {
		f.listeners[i].Send(data)
	}
}

func (f *Feed[T]) Stop() {
	if !atomic.CompareAndSwapUint32(&f.stopped, 0, 1) {
		return
	}
	f.stop()
}

func (f *Feed[T]) stop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	keys := make([]utils.UUID, 0, len(f.listeners))
	for key, wch := range f.listeners {
		wch.Close()
		keys = append(keys, key)
	}

	for _, key := range keys {
		delete(f.listeners, key)
	}
}

func NewFeed[T any]() *Feed[T] {
	f := Feed[T]{}
	f.init()
	return &f
}
