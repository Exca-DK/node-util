package broker

import (
	"context"
	"encoding/json"

	"github.com/Exca-DK/node-util/service/broker"
)

const (
	newNodesQueue  = "nodes.new"
	nodeStatsQueue = "nodes.stats"
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
	PublishNewNodeFound(ctx context.Context, data NewNode) error
	PublishNodeStats(ctx context.Context, data NodeStats) error
}

type wrappedBroker struct {
	broker.Broker
}

// PublishNewNodeFound implements Broker
func (broker wrappedBroker) PublishNewNodeFound(ctx context.Context, data NewNode) error {
	return broker.publish(ctx, data, newNodesQueue)
}

// PublishNodeStats implements Broker
func (broker wrappedBroker) PublishNodeStats(ctx context.Context, data NodeStats) error {
	return broker.publish(ctx, data, nodeStatsQueue)
}

func (broker wrappedBroker) publish(ctx context.Context, data any, topic string) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return broker.PublishMessage(ctx, topic, raw)
}
