package crawl

import "sync/atomic"

func NewSimpleIterator[T any](ch chan T) Iterator[T] {
	return &SimpleIterator[T]{ch: ch, closed: make(chan struct{})}
}

func NewGoIterator[T any](ch chan T) Iterator[T] {
	return &GoIterator[T]{ch: ch, closed: make(chan struct{})}
}

type Iterator[T any] interface {
	Next() T
	Push(t T)
	Stop()
	Stopped() bool
}

type SimpleIterator[T any] struct {
	ch      chan T
	closed  chan struct{}
	stopped uint32
}

func (i *SimpleIterator[T]) Stopped() bool { return atomic.LoadUint32(&i.stopped) == 1 }

func (i *SimpleIterator[T]) Stop() {
	if !atomic.CompareAndSwapUint32(&i.stopped, 0, 1) {
		return
	}
	close(i.closed)
}

func (i SimpleIterator[T]) Push(t T) {
	select {
	case <-i.closed:
	case i.ch <- t:
	}
}

func (i SimpleIterator[T]) Next() T {
	var t T
	select {
	case <-i.closed:
	case t = <-i.ch:
	}
	return t
}

type GoIterator[T any] struct {
	ch      chan T
	closed  chan struct{}
	stopped uint32
}

func (i *GoIterator[T]) Stopped() bool { return atomic.LoadUint32(&i.stopped) == 1 }

func (i *GoIterator[T]) Stop() {
	if !atomic.CompareAndSwapUint32(&i.stopped, 0, 1) {
		return
	}
	close(i.closed)
}

func (i GoIterator[T]) Push(t T) {
	go func() {
		select {
		case <-i.closed:
		case i.ch <- t:
		}
	}()
}

func (i GoIterator[T]) Next() T {
	var t T
	select {
	case <-i.closed:
	case t = <-i.ch:
	}
	return t
}
