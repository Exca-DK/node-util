package v4

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/Exca-DK/node-util/crawler/net/discovery/table"

	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	senderKey uint8 = interfaces.SubProtcolOffset + 1
	enrSeqKey uint8 = senderKey + 1

	BucketSize = 16

	bondExpiration      = 24 * time.Hour
	expiration          = 20 * time.Second
	maxFindnodeFailures = 5
	maxGlobalFailures   = 5
)

func NewHandler(holder interfaces.IdentityHolder, log log.Logger, cfg Config) *handler {
	cfg.validate()
	return &handler{holder: holder, log: log, Config: cfg}
}

// packetHandlerV4 wraps a packet with handler functions.
type packetReactor struct {
	// preverify checks whether the packet is valid and should be handled at all.
	preverify func(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, fromKey v4wire.Pubkey) error
	// handle handles the packet.
	handle func(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, mac []byte) error
}

type handler struct {
	Config
	holder interfaces.IdentityHolder
	tab    table.Tab
	store  *enode.DB
	log    log.Logger

	failureMap sync.Map
}

func (h *handler) Identity() interfaces.IdentityHolder { return h.holder }

func (h *handler) Self() *enode.Node { return h.holder.GetNode() }

func (h *handler) Setup(tab table.Tab, db *enode.DB) {
	h.tab = tab
	h.store = db

	h.log.Info("handler setup finished", log.NewTField("tcp", h.holder.GetTcp()), log.NewTField("udp", h.holder.GetUdp()))
}

// Requests targeted node to respond with pong message.
// Response is being awaited untill deadline.
func (h *handler) MakePingRequest(ctx interfaces.Context, id enode.ID) (uint64, interfaces.RpcInfo) {
	method := v4wire.Ping{}
	result, err := h.makePingRequest(ctx, id, nil)
	return result, interfaces.WrapIntoRpcInfo(err, method.Name())
}

// Requests from targeted node closest nodes to given target.
// It ensures before making actual request that the targeted node can be currently requested.
// The request itself is being awaited within deadline.
func (h *handler) MakeFindNodeRequest(ctx interfaces.Context, id enode.ID, encKey common.EncPubkey) ([]*enode.Node, interfaces.RpcInfo) {
	method := v4wire.Findnode{}
	if !h.tryBond(ctx, id, ctx.Address().(*net.UDPAddr).IP) {
		return nil, interfaces.WrapIntoRpcInfo(ErrBondFailure, method.Name())
	}
	result, err := h.makeFindNodeRequest(ctx, id, encKey)
	return result, interfaces.WrapIntoRpcInfo(err, method.Name())
}

// Requests ENR record from targted node
// It ensures before making actual request that the targeted node can be currently requested.
// The request itself is being awaited within deadline.
func (h *handler) MakeEnrRequest(ctx interfaces.Context, node *enode.Node) (*enode.Node, interfaces.RpcInfo) {
	method := v4wire.ENRRequest{}
	if !h.tryBond(ctx, node.ID(), ctx.Address().(*net.UDPAddr).IP) {
		return nil, interfaces.WrapIntoRpcInfo(ErrBondFailure, method.Name())
	}
	result, err := h.makeEnrRequest(ctx, node)
	return result, interfaces.WrapIntoRpcInfo(err, method.Name())
}

func (h *handler) HandlePacket(ctx interfaces.Context, buf []byte) interfaces.RpcInfo {
	rawpacket, fromKey, hash, err := v4wire.Decode(buf)
	if err != nil {
		return interfaces.WrapIntoRpcInfo(err, "unknown")
	}

	h.log.Trace(fmt.Sprintf("%v<<", rawpacket.Name()), log.NewStringField("id", fromKey.ID().String()), log.NewStringField("addr", ctx.Address().String()))
	reactor, ok := h.selectReactor(rawpacket)
	if !ok {
		return interfaces.WrapIntoRpcInfo(ErrUnknownPacket, "unknown")
	}

	packet := interfaces.NewWirePacket(rawpacket)
	if reactor.preverify != nil {
		if err := reactor.preverify(ctx, packet, fromKey.ID(), fromKey); err != nil {
			return interfaces.WrapIntoRpcInfo(err, rawpacket.Name())
		}
	}
	if reactor.handle != nil {
		return interfaces.WrapIntoRpcInfo(reactor.handle(ctx, packet, fromKey.ID(), hash), rawpacket.Name())
	}
	return interfaces.WrapIntoRpcInfo(nil, rawpacket.Name())
}

func (h *handler) selectReactor(p interfaces.Packet) (packetReactor, bool) {
	var ph packetReactor
	switch p.(type) {
	case *v4wire.Ping:
		ph.preverify = h.VerifyPingRequest
		ph.handle = h.HandlePingRequest
	case *v4wire.Pong:
		ph.preverify = h.VerifyPongResponse
		ph.handle = h.HandlePongResponse
	case *v4wire.Findnode:
		ph.preverify = h.VerifyFindNodeRequest
		ph.handle = h.HandleFindNodeRequest
	case *v4wire.Neighbors:
		ph.preverify = h.VerifyFindNodeResponse
		ph.handle = h.HandleFindNodeResponse
	case *v4wire.ENRResponse:
		ph.preverify = h.verifyEnrResponse
	case *v4wire.ENRRequest:
		ph.preverify = h.verifyENRRequest
		ph.handle = h.handleENRRequest
	default:
		return ph, false
	}
	return ph, true
}

func (h *handler) checkBond(id enode.ID, ip net.IP) bool {
	return time.Since(h.store.LastPongReceived(id, ip)) < bondExpiration
}

// A bond is when both nodes have exchanged ping and pong
// tryBond in that case checks if we have recv. pong from the node and tries to ping if there isn't any recent.
// afterwards it also tries to await within deadline for ping from the targeted node if we haven't got any recent too
// this ensures that the request itself will be accepted by node without any issues
// not checking the recent ping might lead to packet loss because the targeted node will try to bond with us too.
// In that case there is small window for targeted node to ignore out request because it's sent before the node bonds with us.
func (h *handler) tryBond(ctx interfaces.Context, id enode.ID, ip net.IP) bool {
	// if we haven't pinged node recently, try to ping
	if !h.checkBond(id, ip) || h.store.FindFails(id, ip) > maxFindnodeFailures {
		_, err := h.makePingRequest(ctx, id, nil)
		if err != nil {
			return false
		}
	}

	// if there is recent ping its okay to do the request now
	if time.Since(h.store.LastPingReceived(id, ip)) < bondExpiration {
		return true
	}

	time.Sleep(2 * time.Second) //2 seconds is plenty of time to bond
	return time.Since(h.store.LastPingReceived(id, ip)) < bondExpiration
}

// sends packet without awaiting for result
func (h *handler) sendAndForget(ctx interfaces.Context, toId enode.ID, toAddr *net.UDPAddr, packet interfaces.Packet) error {
	return h.send(ctx, false, toId, toAddr, 0, packet, nil, 0)
}

// sends packet with optional awaiting
func (h *handler) send(ctx interfaces.Context, await bool, toId enode.ID, toAddr *net.UDPAddr, exp uint64, packet interfaces.Packet, verify func(interfaces.Response) (bool, bool), exptectedType byte) error {
	enc, _, err := v4wire.Encode(h.holder.GetKey(), packet)
	if err != nil {
		return err
	}
	target := targetablePacket{
		WirePacket: interfaces.NewWirePacket(packet),
		exp:        int64(exp),
		id:         toId,
		ip:         toAddr.IP,
	}

	return ctx.SendRequest(request{
		targetablePacket: target,
		marshalled:       enc,
		verify:           verify,
		expectedType:     exptectedType,
	}, await, toAddr)
}

func (h *handler) isFaulty(id enode.ID) bool {
	faults, ok := h.failureMap.Load(id)
	if !ok {
		return false
	}
	return faults.(uint16) > maxGlobalFailures
}

func (h *handler) markFault(id enode.ID) {
	for {
		val, _ := h.failureMap.LoadOrStore(id, uint16(0))
		if h.failureMap.CompareAndSwap(id, val, val.(uint16)+1) {
			break
		}
	}
}

func (h *handler) clearFault(id enode.ID) {
	h.failureMap.Store(id, uint16(0))
}
