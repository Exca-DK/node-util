package crawl

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Exca-DK/node-util/crawler/net"
	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type QueryCounter struct {
	duration       time.Duration
	LastAttempt    atomic.Int64
	SuccessCounter atomic.Uint32
	FailureCounter atomic.Uint32
}

func (q QueryCounter) String() string {
	return fmt.Sprintf("(succ:%v | fails:%v, last:%v)", q.GetSuccess(), q.GetFailures(), q.GetLastAttempt())
}

// atomically does refresh and return true if successs
func (q *QueryCounter) CanRefresh() bool {
	for {
		lastUnix := q.LastAttempt.Load()
		if time.Since(time.UnixMilli(lastUnix)) < q.duration {
			return false
		}

		if q.LastAttempt.CompareAndSwap(lastUnix, time.Now().Unix()) {
			return true
		}
	}
}

func (q *QueryCounter) GetLastAttempt() time.Time { return time.UnixMilli(q.LastAttempt.Load()) }

func (q *QueryCounter) Bump(failure bool) {
	if failure {
		q.FailureCounter.Add(1)
	} else {
		q.SuccessCounter.Add(1)
	}
	q.LastAttempt.Store(time.Now().UnixMilli())
}

func (q *QueryCounter) GetSuccess() uint32 {
	return q.SuccessCounter.Load()
}

func (q *QueryCounter) GetFailures() uint32 {
	return q.FailureCounter.Load()
}

type Peer struct {
	*common.Node
	info        atomic.Pointer[net.PeerInfo]
	peersAmount atomic.Uint32
	Peers       sync.Map

	infoCounter QueryCounter
	findCounter QueryCounter

	// new keys are being stored here
	mu    sync.Mutex
	fresh []enode.ID

	infoHook func(peer *Peer)
}

func (p *Peer) PeersAmount() int { return int(p.peersAmount.Load()) }

func (p *Peer) GetInfo() (net.PeerInfo, bool) {
	info := p.info.Load()
	if info == nil {
		return net.PeerInfo{}, false
	}
	return *info, true
}

func (p *Peer) UpdateInfo(info net.PeerInfo) bool {
	prev := p.info.Load()
	if prev != nil {
		if reflect.DeepEqual(*prev, info) {
			return false
		}
	}
	p.info.Store(&info)
	if p.infoHook != nil {
		p.infoHook(p)
	}
	return true
}

// marks that such known came from that peer. returns true if its new
func (p *Peer) MarkPeerNode(node *enode.Node) bool {
	_, loaded := p.Peers.LoadOrStore(node.ID(), node)
	if !loaded {
		p.peersAmount.Add(1)
		p.mu.Lock()
		defer p.mu.Unlock()
		p.fresh = append(p.fresh, node.ID())
	}
	return !loaded
}

func (p *Peer) GetPeerNode(id enode.ID) (*enode.Node, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	node, ok := p.Peers.Load(id)
	if !ok {
		return nil, false
	}
	return node.(*enode.Node), true
}

func (p *Peer) GetPeerNodes(ids []enode.ID) []*enode.Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	nodes := make([]*enode.Node, 0, len(ids))
	for _, id := range ids {
		loaded, ok := p.Peers.Load(id)
		if !ok {
			continue
		}
		nodes = append(nodes, loaded.(*enode.Node))
	}
	return nodes
}

func (p *Peer) NewEntries() []enode.ID {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.fresh) == 0 {
		return nil
	}

	nodes := make([]enode.ID, len(p.fresh))
	copy(nodes, p.fresh)
	p.fresh = p.fresh[:0]
	return nodes
}
