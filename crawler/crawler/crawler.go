package crawl

import (
	"context"
	"sync"
	"time"

	"github.com/Exca-DK/node-util/crawler/broker"
	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net"
	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	_enr "github.com/Exca-DK/node-util/crawler/net/enr"
	store "github.com/Exca-DK/node-util/service/db"
	db "github.com/Exca-DK/node-util/service/db/sqlc"

	"github.com/Exca-DK/node-util/crawler/net/discovery"
	"github.com/Exca-DK/node-util/crawler/workers/base"
	"github.com/Exca-DK/node-util/crawler/workers/runner"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	threads = 8

	lookupDuration = 1 * time.Second
	queryDuration  = 5 * time.Second

	infoAttemptFrequency = time.Duration(10 * time.Minute)
	findAttemptFrequency = time.Duration(1 * time.Minute)
)

type Crawler interface {
	AddEnrKey(key string)
	Crawl(ctx context.Context)
	RunFindNodeProc(ctx context.Context)
}

func NewCrawler(handler discovery.IHandler, querier *net.PeerQuierier, broker broker.Broker, store store.Store) *crawler {
	return &crawler{
		handler:  handler,
		buffer:   buffer[*Peer]{},
		logger:   log.NewLogger(),
		keys:     []string{},
		recorder: newRecorder(),
		querier:  querier,
		broker:   broker,
		store:    store,
	}
}

type crawler struct {
	handler discovery.IHandler
	buffer  buffer[*Peer]

	logger log.Logger

	store store.Store

	broker   broker.Broker
	recorder Recorder
	querier  *net.PeerQuierier

	wg sync.WaitGroup

	keys []string
}

func (c *crawler) Wait() {
	c.wg.Wait()
}

func (c *crawler) AddEnrKey(key string) {
	c.keys = append(c.keys, key)
}

func (c *crawler) filterNode(node *enode.Node) bool {
	records := _enr.RetrieveFieldNames(node.Record())
	for _, record := range records {
		for _, key := range c.keys {
			if key == "*" || record == key {
				c.recorder.RecordTargetedNode()
				return true
			}
		}
	}
	c.recorder.RecordUselessNode()
	return false
}

func (c *crawler) Crawl(ctx context.Context) {
	c.wg.Add(1)
	go c.loop(ctx)
	go func() {
		iterator := NewSimpleIterator(make(chan runner.Item[*common.Node]))
		seeker := NewSeeker(c.handler, iterator, c.filterNode)
		seeker.Start(threads)
		defer seeker.Stop()
		go func() {
			<-ctx.Done()
			iterator.Stop()
			c.wg.Done()
		}()
		for {
			node := iterator.Next()
			if iterator.Stopped() {
				break
			}

			c.recorder.RecordRecvNode()
			peer := &Peer{Node: node.GetData(),
				infoCounter: QueryCounter{duration: infoAttemptFrequency},
				findCounter: QueryCounter{duration: findAttemptFrequency},
				infoHook: func(peer *Peer) {
					info, _ := peer.GetInfo()
					c.broker.PublishNewNodeFound(ctx, broker.NewNode{
						ENR:     peer.URLv4(),
						Enrs:    _enr.RetrieveFieldNames(peer.Record()),
						Name:    info.Name,
						Version: info.Version,
						Caps:    info.Caps,
						Extras:  info.Extra,
					})
					c.store.CreateRlpRecord(ctx, db.CreateRlpRecordParams{ID: peer.URLv4(), RawName: info.Name, ProtocolVersion: int32(info.Version), Caps: info.Caps})
				},
			}

			id := peer.URLv4()
			fields := _enr.RetrieveFieldNames(peer.Record())
			c.store.CreateEnrRecord(ctx, db.CreateEnrRecordParams{ID: id, Fields: fields})
			// Split publish in two parts. first publish info that its just been discovered
			// later down the chain if we manage to connect through rlpx, publish once again with full info
			c.broker.PublishNewNodeFound(ctx, broker.NewNode{ENR: id, Enrs: fields})
			c.buffer.Enqueue(peer)
		}
	}()
}

func (c *crawler) loop(ctx context.Context) {
	iterator := NewSimpleIterator(make(chan runner.Item[*Peer]))
	defer iterator.Stop()
	r := runner.NewReader[runner.Item[*Peer]](threads, iterator, func(feed base.JobStats) {})
	r.Start(newCrawlerProc(log.NewLogger(), threads, c.handler, c.querier))
	defer r.Stop()

	lookupTicker := time.NewTicker(lookupDuration)
	defer lookupTicker.Stop()
	for {
		select {
		case <-lookupTicker.C:
			node, ok := c.buffer.Dequeue()
			if !ok {
				continue
			}
			c.buffer.Enqueue(node)
			if node.findCounter.CanRefresh() {
				iterator.Push(runner.NewItem(node))
				c.recorder.RecordLookup()
			}
			entries := node.NewEntries()
			if len(entries) == 0 {
				continue
			}

			enodes := node.GetPeerNodes(entries)
			findTotalCount := node.findCounter.GetFailures() + node.findCounter.GetSuccess()
			infoTotalCount := node.infoCounter.GetFailures() + node.infoCounter.GetSuccess()
			nodes := make([]string, len(enodes))
			for i, enode := range enodes {
				nodes[i] = enode.URLv4()
			}
			c.broker.PublishNodeStats(ctx, broker.NodeStats{
				ENR:      node.URLv4(),
				Queried:  int(findTotalCount + infoTotalCount),
				Nodes:    node.PeersAmount(),
				NewNodes: nodes,
			})

			info, _ := node.GetInfo()
			c.logger.Info("peer stats",
				log.NewTField("info", info),
				log.NewTField("peers", node.PeersAmount()),
				log.NewTField("newPeers", len(entries)),
				log.NewTField("findStats", node.findCounter.String()),
				log.NewTField("infoStats", node.infoCounter.String()),
				log.NewStringField("ip", node.IP().String()),
				log.NewStringField("id", node.ID().String()))
		case <-ctx.Done():
			return
		}
	}
}
