package events

import (
	"sync"
	"testing"
)

func TestFeedSendRecv(t *testing.T) {
	t.Parallel()
	var (
		loops     = 100000
		listeners = []chan uint32{make(chan uint32, 1000), make(chan uint32, 1000), make(chan uint32, 1000), make(chan uint32, 1000)}
		subs      = make([]Subscription, len(listeners))
	)

	feed := NewFeed[uint32]()

	for i, ch := range listeners {
		sub := feed.Subscribe(ch)
		subs[i] = sub
	}

	var wg sync.WaitGroup
	wg.Add(loops * len(listeners))
	for _, listener := range listeners {
		go func(l chan uint32) {
			for {
				<-l
				wg.Done()
			}
		}(listener)
	}

	for i := 0; i < loops; i++ {
		feed.Send(uint32(i) + 1)
	}

	wg.Wait()
}

func TestFeedUnsubscribe(t *testing.T) {
	t.Parallel()
	var (
		loops     = 100000
		listeners = []chan uint32{make(chan uint32, 0), make(chan uint32, 0), make(chan uint32, 0), make(chan uint32, 0)}
		subs      = make([]Subscription, len(listeners))
	)

	feed := NewFeed[uint32]()

	for i, ch := range listeners {
		sub := feed.Subscribe(ch)
		subs[i] = sub
	}

	var wg sync.WaitGroup
	wg.Add(len(listeners))
	for _, listener := range listeners {
		go func(l chan uint32) {
			for {
				_, ok := <-l
				if !ok {
					wg.Done()
					return
				}
			}
		}(listener)
	}

	for i := 0; i < loops; i++ {
		feed.Send(uint32(i) + 1)
		if i == 100 {
			for _, sub := range subs {
				sub.Unsubscribe()
			}
		}
	}
	wg.Wait()
}
