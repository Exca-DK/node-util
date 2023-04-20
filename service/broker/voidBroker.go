package broker

import (
	"context"
)

func NewVoidBroker() Broker {
	return voidBroker{}
}

// VoidBroker is an lazy way to not use actual broke.
// Messages are discarded and subscriptions are empty
type voidBroker struct{}

func (b voidBroker) PublishMessage(ctx context.Context, topic string, data []byte) error {
	return nil
}
func (b voidBroker) SubscribeTopic(topic string) (chan BrokerMessage, Subscription, error) {
	return make(chan BrokerMessage), newSubsciption(voidSubscription{}), nil
}

func (b voidBroker) Stop(err error) {}
