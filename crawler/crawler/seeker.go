package crawl

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Exca-DK/node-util/crawler/events"
	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery"
	"github.com/Exca-DK/node-util/crawler/net/discovery/common"

	"github.com/Exca-DK/node-util/crawler/workers/runner"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type handler interface {
	SubscribeNodeEvents() (<-chan discovery.NodeEvent, events.Subscription)
	RequestENR(n *enode.Node) (*enode.Node, error)
}

func NewSeeker(handler handler, dst Iterator[runner.Item[*common.Node]], filter nodeFilter) *seeker {
	return &seeker{
		handler:      handler,
		seen:         map[enode.ID]struct{}{},
		filter:       filter,
		dstIterator:  dst,
		recvIterator: NewSimpleIterator(make(chan runner.Item[*enode.Node])),
		log:          log.NewLogger(),
		sig:          make(chan struct{}),
		done:         make(chan struct{}),
	}
}

type nodeFilter func(enrKeys *enode.Node) bool

type seeker struct {
	seen map[enode.ID]struct{}

	log log.Logger

	filter  nodeFilter
	handler handler

	recvIterator Iterator[runner.Item[*enode.Node]]
	dstIterator  Iterator[runner.Item[*common.Node]]

	runner runner.Runner[runner.Item[*enode.Node], runner.Item[*common.Node]]

	sig     chan struct{}
	done    chan struct{}
	running atomic.Bool
	wg      sync.WaitGroup
}

func (s *seeker) Start(threads int) {
	if !s.running.CompareAndSwap(false, true) {
		return
	}
	if s.filter == nil {
		s.filter = func(enrKeys *enode.Node) bool { return true }
	}
	s.wg.Add(1)
	s.runner = runner.NewReaderWriter[runner.Item[*enode.Node], runner.Item[*common.Node]](threads, s.recvIterator, s.dstIterator, nil)
	s.runner.Start(newSeekerProc(log.NewLogger(), threads, s.filter, s.handler.RequestENR))
	go s.start(threads)
}

func (s *seeker) start(threads int) {
	defer s.cleanup()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.eventLoop(ctx)

	<-s.sig
	close(s.done)
}

func (s *seeker) cleanup() {

	s.log.Debug("cleanup starting")
	s.dstIterator.Stop()
	s.recvIterator.Stop()
	s.runner.Stop()
	s.runner.Wait()
	s.log.Debug("cleanup finished")
	s.wg.Done()
}

func (s *seeker) Stop() {
	select {
	case <-s.done:
		return
	case s.sig <- struct{}{}:
	}

	s.wg.Wait()
}

func (s *seeker) RegisterFilter(f func(enr *enode.Node) bool) {
	if s.running.Load() {
		return
	}
	s.filter = f
}

func (s *seeker) eventLoop(ctx context.Context) {
	ch, sub := s.handler.SubscribeNodeEvents()
	defer sub.Unsubscribe()
OUTER:
	for {
		select {
		case event := <-ch:
			if event.Type != discovery.SEEN {
				continue
			}

			if _, ok := s.seen[event.GetNode().ID()]; ok {
				continue
			}
			s.seen[event.GetNode().ID()] = struct{}{}
			s.recvIterator.Push(runner.NewItem(event.GetNode()))
		case <-ctx.Done():
			break OUTER
		}
	}
}
