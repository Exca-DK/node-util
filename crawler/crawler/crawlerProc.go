package crawl

import (
	"reflect"
	"sync"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net"
	"github.com/Exca-DK/node-util/crawler/net/discovery"
	"github.com/Exca-DK/node-util/crawler/workers/runner"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type crawlerProc struct {
	logger  log.Logger
	threads int

	querier *net.PeerQuierier

	handler discovery.RequestHandler
	name    string
}

func (t *crawlerProc) TargetedThreads() int {
	return t.threads
}

func (t *crawlerProc) Identity() string {
	return t.name
}

func (t *crawlerProc) OnData(item runner.Item[*Peer]) (runner.NOP, error) {
	src := item.GetData()

	if src.infoCounter.CanRefresh() {
		info, err := t.querier.Query(src.GetNode())
		if err == nil {
			updated := src.UpdateInfo(info)
			if updated {
				t.logger.Info("updated info",
					log.NewStringField("ip", src.IP().String()),
					log.NewStringField("id", src.ID().String()),
					log.NewTField("info", info))
			}
			src.infoCounter.Bump(!updated)
		} else {
			src.infoCounter.Bump(true)
		}
	}

	nodes, err := t.handler.AskForRandomNodes(src.GetNode())
	src.findCounter.Bump(err != nil)
	if err != nil {
		src.findCounter.Bump(true)
		return runner.NOP{}, err
	}
	var wg sync.WaitGroup
	for i := range nodes {
		node := nodes[i]
		if !src.MarkPeerNode(node) {
			continue
		}

		wg.Add(1)
		go func(node *enode.Node) {
			t.handler.Ping(node)
			wg.Done()
		}(node)
	}
	wg.Wait()
	return runner.NOP{}, nil
}

func newCrawlerProc(log log.Logger, threads int, handler discovery.RequestHandler, querier *net.PeerQuierier) *crawlerProc {
	streamer := crawlerProc{}
	streamer.name = reflect.TypeOf(streamer).Name()
	streamer.logger = log
	streamer.threads = threads
	streamer.handler = handler
	streamer.querier = querier
	return &streamer
}
