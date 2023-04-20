package discovery

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/Exca-DK/node-util/crawler/net/sims"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestLookupQueryDepthOne(t *testing.T) {
	t.Parallel()
	doTestQuery(t, 1, 16, 3)
}

func TestLookupQueryDepthTwo(t *testing.T) {
	t.Parallel()
	doTestQuery(t, 2, 16, 3)
}

func TestLookupQueryDepthThree(t *testing.T) {
	t.Parallel()
	doTestQuery(t, 3, 16, 3)
}

func doTestQuery(t *testing.T, depth int, want int, alpha int) {
	bootnodes := make([]*enode.Node, want)
	for i := 0; i < want; i++ {
		a, _ := sims.NewUdpPipe()
		bootnodes[i] = enode.NewV4(&newkey().PublicKey, a.LocalIp(), a.LocalPort(), a.LocalPort())
	}

	randomNodes := make([]*enode.Node, 0, want*depth)
	for i := 0; i < ((want * depth * alpha) / 2); i++ {
		a, b := sims.NewUdpPipe()
		randomNodes = append(randomNodes, enode.NewV4(&newkey().PublicKey, a.LocalIp(), a.LocalPort(), a.LocalPort()))
		randomNodes = append(randomNodes, enode.NewV4(&newkey().PublicKey, b.LocalIp(), b.LocalPort(), b.LocalPort()))
	}

	entries := common.NodesByDistance{Target: enode.NewV4(&newkey().PublicKey, net.IPv4(127, 0, 0, 1), 5050, 5050).ID()}
	for _, random := range bootnodes {
		entries.Push(common.WrapNode(random), len(bootnodes)+len(randomNodes))
	}
	for _, random := range randomNodes {
		entries.Push(common.WrapNode(random), len(bootnodes)+len(randomNodes))
	}

	var mu sync.Mutex
	counter := 0
	offset := len(entries.Entries()) / (alpha * depth)
	remainder := len(entries.Entries()) % (alpha * depth)
	quertyFunc := func(n *enode.Node) ([]*enode.Node, error) {
		mu.Lock()
		defer mu.Unlock()
		nodes := entries.Enodes()
		nodeSlice := nodes[len(nodes)-offset*(counter+1) : len(nodes)-offset*counter]
		counter++
		return nodeSlice, nil
	}

	lookup := newLookup(entries.Enodes()[len(entries.Entries())-want:], quertyFunc, entries.Target, depth, want, alpha)
	if err := lookup.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	result := lookup.Result()
	for i := 0; i < len(result); i++ {
		a, b := result[i], entries.Enodes()[i+remainder]
		if a.ID() != b.ID() {
			t.Fatal("node mismatch")
		}
	}
}
