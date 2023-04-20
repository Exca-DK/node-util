package discovery

import (
	"context"
	"crypto/rand"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/Exca-DK/node-util/crawler/events"
	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery/table"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
)

var (
	errUnknownPacket = errors.New("unknown packet")
	errUnsupported   = errors.New("unsupported")
)

type EventType table.HookType

const (
	ADDED   EventType = EventType(table.Added)
	SEEN    EventType = EventType(table.Seen)
	REMOVED EventType = EventType(table.Removed)
)

func (ev EventType) String() string {
	switch ev {
	case ADDED:
		return "added"
	case SEEN:
		return "seen"
	case REMOVED:
		return "removed"
	default:
		return "unknown"
	}
}

type NodeEvent struct {
	interfaces.SeenNode
	Type EventType
}

type Config struct {
	Bootnodes    []*enode.Node
	RunTab       bool
	IgnoreLimits bool
}

func (c *Config) Verify() {
	if c.Bootnodes == nil {
		c.Bootnodes = make([]*enode.Node, 0)
	}
}

func NewHandler(requester Writer, sub SubHandler, db *enode.DB, cfg Config) (*PacketHandler, error) {
	cfg.Verify()
	h := &PacketHandler{
		Config:    cfg,
		sub:       sub,
		awaiter:   NewAwaiter(log.NewLogger()),
		requester: requester,

		ctxPool: sync.Pool{
			New: func() interface{} {
				return new(ctx)
			},
		},
		recorder: newRecorder(),
		feed:     events.NewFeed[NodeEvent](),
	}

	tab, err := table.NewTable(h, db, cfg.Bootnodes, log.NewLogger(), cfg.IgnoreLimits, func(sn interfaces.SeenNode, ht table.HookType) {
		h.feed.Send(NodeEvent{SeenNode: sn, Type: EventType(ht)})
	})
	if err != nil {
		return nil, err
	}

	h.table = tab
	sub.Setup(tab, db)
	return h, nil
}

type SubHandler interface {
	Identity() interfaces.IdentityHolder
	Self() *enode.Node
	Setup(tab table.Tab, db *enode.DB)
	HandlePacket(ctx interfaces.Context, buf []byte) interfaces.RpcInfo
	MakePingRequest(ctx interfaces.Context, id enode.ID) (uint64, interfaces.RpcInfo)
	MakeFindNodeRequest(ctx interfaces.Context, id enode.ID, target common.EncPubkey) ([]*enode.Node, interfaces.RpcInfo)
	MakeEnrRequest(ctx interfaces.Context, node *enode.Node) (*enode.Node, interfaces.RpcInfo)
}

type LookupHandler interface {
	LookupRandom() ([]*enode.Node, error)
	LookupSelf() ([]*enode.Node, error)
}

type RequestHandler interface {
	AskForRandomNodes(node *enode.Node) ([]*enode.Node, error)
	Ping(n *enode.Node) (uint64, error)
	RequestENR(n *enode.Node) (*enode.Node, error)
}

type IHandler interface {
	LookupHandler
	RequestHandler
	Run()
	Self() *enode.Node
	SubscribeNodeEvents() (<-chan NodeEvent, events.Subscription)
	HandlePacket(buf []byte, from *net.UDPAddr) error
}

type PacketKind uint8

type PacketHandler struct {
	Config
	sub       SubHandler
	awaiter   *awaiter
	requester Writer

	feed *events.Feed[NodeEvent]

	table *table.Table

	recorder Recorder

	ctxPool sync.Pool
}

func (h *PacketHandler) Run() {
	if h.RunTab {
		go h.table.Run()
	}
}

func (h *PacketHandler) HandlePacket(buf []byte, from *net.UDPAddr) error {
	ctx := h.getCtx(from)
	defer h.ctxPool.Put(ctx)
	ctx.ip = from
	ctx.send = h.execute
	ctx.match = h.awaiter.MatchResponse

	f := h.record(true)
	result := h.sub.HandlePacket(ctx, buf)
	f(result)
	return result.Error()
}

func (h *PacketHandler) execute(request interfaces.Request, await bool, addr net.Addr) error {
	var awaitable interfaces.AwaitableRequest
	if await {
		awaitable = h.awaiter.AddPending(request, request.Verify)
	}

	if err := h.requester.write(addr.(*net.UDPAddr), request); err != nil {
		return err
	}

	if awaitable != nil {
		awaitable.Await()
		return awaitable.Error()
	}

	return nil
}

func (h *PacketHandler) Ping(n *enode.Node) (uint64, error) {
	ctx := h.getCtx(&net.UDPAddr{IP: n.IP(), Port: n.UDP()})
	defer h.ctxPool.Put(ctx)
	f := h.record(false)
	result, info := h.sub.MakePingRequest(ctx, n.ID())
	f(info)
	return result, info.Error()
}

func (h *PacketHandler) AskForRandomNodes(node *enode.Node) ([]*enode.Node, error) {
	var target common.EncPubkey
	rand.Read(target[:])
	return h.runNodesQuery(&net.UDPAddr{IP: node.IP(), Port: node.UDP()}, node.ID(), target)
}

func (h *PacketHandler) LookupRandom() ([]*enode.Node, error) {
	var target common.EncPubkey
	rand.Read(target[:])

	queryFunc := func(n *enode.Node) ([]*enode.Node, error) {
		return h.runNodesQuery(&net.UDPAddr{IP: n.IP(), Port: n.UDP()}, n.ID(), target)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodes := h.table.FindnodeByID(target.Id(), 16, true)
	if len(nodes.Entries()) == 0 {
		return nil, errors.New("seed failure")
	}

	lookup := newLookup(nodes.Enodes(), queryFunc, target.Id(), 3, 16, 3)
	if err := lookup.Run(ctx); err != nil {
		return nil, err
	}

	return lookup.result.Enodes(), nil
}

func (h *PacketHandler) runNodesQuery(address *net.UDPAddr, id enode.ID, target common.EncPubkey) ([]*enode.Node, error) {
	ctx := h.getCtx(address)
	defer h.ctxPool.Put(ctx)
	f := h.record(false)
	nodes, info := h.sub.MakeFindNodeRequest(ctx, id, target)
	f(info)
	return nodes, info.Error()
}

func (h *PacketHandler) RequestENR(n *enode.Node) (*enode.Node, error) {
	ctx := h.getCtx(&net.UDPAddr{IP: n.IP(), Port: n.UDP()})
	defer h.ctxPool.Put(ctx)
	f := h.record(false)
	nodes, info := h.sub.MakeEnrRequest(ctx, n)
	f(info)
	return nodes, info.Error()
}

func (h *PacketHandler) Self() *enode.Node {
	return h.sub.Self()
}

func (h *PacketHandler) LookupSelf() ([]*enode.Node, error) {
	return nil, errors.New("not implemented")
}

// helper method to retrieve new usable context
func (h *PacketHandler) getCtx(address *net.UDPAddr) *ctx {
	ctx := h.ctxPool.Get().(*ctx)
	ctx.ip = address
	ctx.send = h.execute
	ctx.match = h.awaiter.MatchResponse
	return ctx
}

func (h *PacketHandler) SubscribeNodeEvents() (<-chan NodeEvent, events.Subscription) {
	ch := make(chan NodeEvent)
	return ch, h.feed.Subscribe(ch)
}

func (h *PacketHandler) record(ext bool) func(interfaces.RpcInfo) {
	ts := time.Now()
	return func(result interfaces.RpcInfo) {
		method := result.Packet()
		var code int
		if result.Error() == nil {
			code = 200
		} else {
			code = 404
		}
		h.recorder.RecordRpcDuration(time.Since(ts), method, code, ext)
		h.recorder.RecordRpcRequest(method, code, ext)
	}
}
