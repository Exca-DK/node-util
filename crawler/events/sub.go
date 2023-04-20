package events

import (
	"github.com/Exca-DK/node-util/crawler/utils"
)

type UnsubcribeFunc func(utils.UUID)

type Subscription interface{ Unsubscribe() }

type emptySubscription struct{}

func (s *emptySubscription) Unsubscribe() {}

type subscription struct {
	uuid utils.UUID
	f    UnsubcribeFunc
}

func (s *subscription) Unsubscribe() {
	s.f(s.uuid)
}
