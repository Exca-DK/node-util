package crawl

import (
	"errors"
	"reflect"
	"time"

	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/Exca-DK/node-util/crawler/workers/runner"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Exca-DK/node-util/crawler/log"
)

var (
	errNotInterested = errors.New("not interested")
)

type seekerProcessor struct {
	logger log.Logger

	threads    int
	enrFunc    func(*enode.Node) (*enode.Node, error)
	verifyFunc func(*enode.Node) bool

	name string
}

func (t *seekerProcessor) TargetedThreads() int {
	return t.threads
}

func (t *seekerProcessor) Identity() string {
	return t.name
}

func (t *seekerProcessor) OnData(item runner.Item[*enode.Node]) (runner.Item[*common.Node], error) {
	var err error
	var enrNode *enode.Node

	for i := 0; i < 3; i++ {
		enrNode, err = t.enrFunc(item.GetData())
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return runner.Item[*common.Node]{}, err
	}

	if !t.verifyFunc(enrNode) {
		return runner.Item[*common.Node]{}, errNotInterested
	}

	wrapped := common.WrapNode(enrNode)
	wrapped.SetTimestamp(time.Now())
	wrapped.Bump()
	return runner.NewItem(wrapped), nil
}

func newSeekerProc(log log.Logger, threads int, verify func(*enode.Node) bool, getEnr func(*enode.Node) (*enode.Node, error)) *seekerProcessor {
	streamer := seekerProcessor{}
	streamer.name = reflect.TypeOf(streamer).Name()
	streamer.logger = log
	streamer.threads = threads
	streamer.enrFunc = getEnr
	streamer.verifyFunc = verify
	return &streamer
}
