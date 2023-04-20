package crawl

import (
	"net"
	"testing"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery"
	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/Exca-DK/node-util/crawler/workers/runner"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestSeeker(t *testing.T) {
	var (
		amount = 1001
	)
	h := discovery.NewMockHandler(discovery.NewRandomIdentity("", net.IPv4(127, 0, 0, 1), 12345, 12345, log.NewLogger()).GetNode())
	iterator := NewSimpleIterator(make(chan runner.Item[*common.Node]))
	s := NewSeeker(h, iterator, func(enrKeys *enode.Node) bool { return true })

	s.Start(4)
	defer s.Stop()

	nodes := make(map[enode.ID]*enode.Node)
	for i := 0; i < amount; i++ {
		n := discovery.NewRandomIdentity("", net.IPv4(127, 0, 0, 1), 12345, 12345, log.NewLogger()).GetNode()
		nodes[n.ID()] = n
	}

	go func() {
		for _, node := range nodes {
			h.AddToNodeDb(node.ID(), []*enode.Node{node})
		}
	}()

	unique := make(map[enode.ID]struct{}, len(nodes))
	for i := 0; i < amount; i++ {
		item := iterator.Next().GetData()
		if _, ok := unique[item.ID()]; ok {
			t.Fatal("got non-unique??")
		}

		ref, ok := nodes[item.ID()]

		if !ok {
			t.Fatalf("missing node. got: %v", item.ID())
		}

		if ref.ID() != item.ID() {
			t.Fatalf("node mismatch. got: %v, wants: %v", item.ID(), ref.ID())
		}

		unique[item.ID()] = struct{}{}
	}
	iterator.Stop()
}
