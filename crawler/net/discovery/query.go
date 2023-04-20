package discovery

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Exca-DK/node-util/crawler/net/discovery/common"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func newLookup(seedNodes []*enode.Node, query queryFunc, target enode.ID, depth int, wants int, alpha int) *lookup {
	l := lookup{
		query:  query,
		ch:     make(chan []*enode.Node),
		asked:  map[enode.ID]struct{}{},
		seen:   map[enode.ID]struct{}{},
		buffer: queryBuffer{buffer: make([]*enode.Node, 0, wants*alpha)},
		result: common.NodesByDistance{Target: target},
		depth:  depth,
		wants:  wants,
		alpha:  alpha,
	}
	for _, node := range seedNodes {
		l.result.Push(common.WrapNode(node), wants)
		l.seen[node.ID()] = struct{}{}
	}
	return &l
}

type queryFunc func(*enode.Node) ([]*enode.Node, error)

type lookup struct {
	query queryFunc

	ch chan []*enode.Node

	asked, seen map[enode.ID]struct{}

	buffer queryBuffer
	result common.NodesByDistance

	depth int
	wants int
	alpha int
}

func (it *lookup) Result() []*enode.Node {
	return it.result.Enodes()
}

func (it *lookup) Run(ctx context.Context) error {
	for it.queryNewNodes(ctx, it.result.Enodes()) {
		nodes := it.buffer.peek()
		for i := 0; i < len(nodes); i++ {
			node := nodes[i]
			_, ok := it.seen[node.ID()]
			if ok {
				continue
			}
			it.seen[node.ID()] = struct{}{}
			it.result.Push(common.WrapNode(node), it.wants)
		}
		it.buffer.slide()
	}
	return nil
}

// query new nodes from buffer. It tries to acquire K succefull responses
// return false if none succesfull queries has been performed or if the query depth would exceed
func (it *lookup) queryNewNodes(ctx context.Context, buffer []*enode.Node) bool {
	if it.depth == 0 {
		return false
	}
	it.depth--
	slots := make(chan struct{}, it.alpha)
	for i := 0; i < it.alpha; i++ {
		slots <- struct{}{}
	}
	var doneCounter atomic.Uint32
	var wg sync.WaitGroup
	// try to query alpha closest nodes
	// try further node each time node fails to deliver unknown nodes
	for i := 0; i < len(buffer); i++ {
		node := buffer[i]
		if !it.canAsk(node) {
			continue
		}

		select {
		case <-slots:
		case <-ctx.Done():
			return false
		}
		if doneCounter.Load() == uint32(it.alpha) {
			break
		}
		wg.Add(1)
		go func(n *enode.Node) {
			defer wg.Done()
			nodes, err := it.query(n)
			if err != nil {
				slots <- struct{}{}
				return
			}
			added := it.buffer.append(nodes)
			if added > 0 {
				new := doneCounter.Add(1)
				if new == uint32(it.alpha) {
					close(slots)
				}
				return
			}
			slots <- struct{}{}
		}(node)
	}
	wg.Wait() //there might be pending request after loop execution. wait for it in that case
	return doneCounter.Load() != 0
}

// checks if its possible to query provided node and marks as asked if true
func (it *lookup) canAsk(node *enode.Node) bool {
	_, ok := it.asked[node.ID()]
	if ok {
		return false
	}
	it.asked[node.ID()] = struct{}{}
	return true
}

type queryBuffer struct {
	buffer []*enode.Node
	mu     sync.Mutex
}

// appends unique elements to buffer
// safe for concurrent use, returns amount of appended nodes
func (buff *queryBuffer) append(nodes []*enode.Node) int {
	buff.mu.Lock()
	defer buff.mu.Unlock()
	added := 0
	for _, node := range nodes {
		var exists bool
		for _, curr := range buff.buffer {
			if curr.ID() == node.ID() {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		buff.buffer = append(buff.buffer, node)
		added++
	}
	return added
}

// retrieves new entries. not safe
func (buff *queryBuffer) peek() []*enode.Node {
	if len(buff.buffer) == 0 {
		return nil
	}

	return buff.buffer
}

func (buff *queryBuffer) slide() {
	//clear
	for i := 0; i < len(buff.buffer); i++ {
		buff.buffer[i] = nil
	}
	buff.buffer = buff.buffer[:0]
}
