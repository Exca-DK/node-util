package broker

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/event"
)

type voidSubscription struct{}

func (sub voidSubscription) Err() <-chan error { return make(<-chan error) }
func (sub voidSubscription) Unsubscribe()      {}

func newSubsciption(sub event.Subscription) Subscription { return &subscription{Subscription: sub} }

type Subscription interface {
	Unsubscribe()
}

type subscription struct {
	event.Subscription
	called atomic.Bool
}

func (sub *subscription) Unsubscribe() {
	if !sub.called.CompareAndSwap(false, true) {
		return
	}
	sub.Subscription.Unsubscribe()
}

type Feed[T any] struct {
	feed      event.FeedOf[T]
	listeners atomic.Int32
	callback  atomic.Value
	delta     atomic.Int32
}

func (dfeed *Feed[T]) Unsubscribe() {
	dfeed.listeners.Add(-1)
	call := dfeed.callback.Load()
	if call == nil {
		dfeed.delta.Add(-1)
		return
	}
	call.(func(int))(-1)
}

func (dfeed *Feed[T]) Subscribe(ch chan T) Subscription {
	call := dfeed.callback.Load()
	if call != nil {
		call.(func(int))(1 + int(dfeed.delta.Load()))
	} else {
		dfeed.delta.Add(1)
	}
	dfeed.listeners.Add(1)
	return newSubsciption(dfeed.feed.Subscribe(ch))
}

func (dfeed *Feed[T]) Notify(msg T) {
	dfeed.feed.Send(msg)
}
