package runner

import "time"

type Item[T any] struct {
	ts   int64
	data T
}

func (i Item[T]) GetData() T {
	return i.data
}

func (i Item[T]) CreationTime() time.Time {
	return time.UnixMicro(i.ts)
}

func NewItem[T any](value T) Item[T] {
	return Item[T]{
		ts:   time.Now().UnixMicro(),
		data: value,
	}
}

func NewItemWithTs[T any](value T, ts time.Time) Item[T] {
	return Item[T]{
		ts:   ts.UnixMicro(),
		data: value,
	}
}
