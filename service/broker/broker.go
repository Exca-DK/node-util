package broker

import (
	"context"
	"errors"
)

var (
	ErrBrokerStopped = errors.New("broker stopped")
)

type BrokerMessage struct {
	Data []byte
}

type Broker interface {
	PublishMessage(ctx context.Context, topic string, data []byte) error
	SubscribeTopic(topic string) (chan BrokerMessage, Subscription, error)
	Stop(err error)
}
