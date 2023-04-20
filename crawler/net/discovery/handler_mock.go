package discovery

import (
	"sync"
	"sync/atomic"

	"github.com/Exca-DK/node-util/crawler/events"
	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func NewMockHandler(n *enode.Node) *MockHandler {
	return &MockHandler{
		n:       n,
		feed:    events.NewFeed[NodeEvent](),
		pings:   map[enode.ID]*mockPing{},
		nodesDb: make(map[enode.ID]map[enode.ID]*enode.Node),
		nodes:   map[enode.ID]*enode.Node{},
	}
}

type MockHandler struct {
	n    *enode.Node
	feed *events.Feed[NodeEvent]

	mu      sync.Mutex //everything below is secured by mutex
	pings   map[enode.ID]*mockPing
	nodesDb map[enode.ID]map[enode.ID]*enode.Node
	nodes   map[enode.ID]*enode.Node //flattened nodesDb
}

func (h *MockHandler) SetPingSlot(id enode.ID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pings[id] = &mockPing{}

}
func (h *MockHandler) AddToNodeDb(id enode.ID, nodes []*enode.Node) {
	h.mu.Lock()
	defer h.mu.Unlock()
	db, ok := h.nodesDb[id]
	if !ok {
		db = make(map[enode.ID]*enode.Node)
	}
	for _, node := range nodes {
		h.feed.Send(NodeEvent{SeenNode: common.WrapNode(node), Type: SEEN})
		db[node.ID()] = node
		_, ok := h.nodes[node.ID()]
		if !ok {
			h.nodes[node.ID()] = node
			h.feed.Send(NodeEvent{SeenNode: common.WrapNode(node), Type: ADDED})
		}
	}
	h.nodesDb[id] = db
}

func (h *MockHandler) SubscribeNodeEvents() (<-chan NodeEvent, events.Subscription) {
	ch := make(chan NodeEvent)
	return ch, h.feed.Subscribe(ch)
}

func (h *MockHandler) LookupSelf() ([]*enode.Node, error) {
	return h.LookupRandom()
}

func (h *MockHandler) Self() *enode.Node {
	return h.n
}

func (h *MockHandler) RequestENR(n *enode.Node) (*enode.Node, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	v, ok := h.nodes[n.ID()]
	if !ok {
		return nil, errRespDeadlinePassed
	}
	return v, nil
}

func (h *MockHandler) LookupRandom() ([]*enode.Node, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	nodes := make([]*enode.Node, 0, 16)
	c := 0
	for key := range h.nodes {
		if c == 16 {
			break
		}
		c++
		h.feed.Send(NodeEvent{SeenNode: common.WrapNode(h.nodes[key]), Type: SEEN})
		nodes = append(nodes, h.nodes[key])
	}

	if len(nodes) == 0 {
		return nil, errRespDeadlinePassed
	}
	return nodes, nil
}

type mockPing struct {
	v atomic.Uint64
}

func (h *MockHandler) Ping(n *enode.Node) (uint64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	val, ok := h.pings[n.ID()]
	if !ok {
		return 0, errRespDeadlinePassed
	}
	return val.v.Add(1), nil
}

func (h *MockHandler) AskForRandomNodes(node *enode.Node) ([]*enode.Node, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	v, ok := h.nodesDb[node.ID()]
	if !ok {
		return nil, errRespDeadlinePassed
	}

	nodes := make([]*enode.Node, 0, 16)
	c := 0
	for id := range v {
		if c == 16 {
			break
		}
		c++
		h.feed.Send(NodeEvent{SeenNode: common.WrapNode(v[id]), Type: SEEN})
		nodes = append(nodes, v[id])
	}

	if len(nodes) == 0 {
		return nil, errRespDeadlinePassed
	}
	return nodes, nil
}
