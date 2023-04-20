package broker

import (
	"context"
	"encoding/json"

	"github.com/Exca-DK/node-util/service/broker"
)

const (
	newNodesQueue = "nodes.new"
)

type NewNode struct {
	//discovery
	ENR  string
	Enrs []string

	//rlpx
	Name    string
	Version int
	Caps    []string
	Extras  map[string]string
}

type NodeStats struct {
	ENR      string
	Queried  int
	Nodes    int
	NewNodes []string
}

// wraps generic broker into concrete one
func NewBroker(broker broker.Broker) Broker {
	return wrappedBroker{Broker: broker}
}

type Broker interface {
	SubscribeNewNodes(ctx context.Context) (chan NewNode, broker.Subscription, error)
}

type wrappedBroker struct {
	broker.Broker
}

// PublishNodeStats implements Broker
func (broker wrappedBroker) SubscribeNewNodes(ctx context.Context) (chan NewNode, broker.Subscription, error) {
	ch, sub, err := broker.Broker.SubscribeTopic(newNodesQueue)
	if err != nil {
		return nil, nil, err
	}

	nodeCh := make(chan NewNode)
	go func() {
		for {
			msg, ok := <-ch
			if !ok {
				return
			}
			// TODO  use protobuf
			obj := new(NewNode)
			if err := json.Unmarshal(msg.Data, obj); err != nil {
				continue
			}
			nodeCh <- *obj
		}
	}()
	return nodeCh, sub, nil
}
