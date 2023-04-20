package v4

import (
	"crypto/ecdsa"
	"net"
	"time"

	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/netutil"
)

func (h *handler) VerifyFindNodeRequest(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, fromKey v4wire.Pubkey) error {
	if !h.Config.ResponseFindNode {
		return ErrHandlerDisabled
	}
	req, ok := p.GetSource().(*v4wire.Findnode)
	if !ok {
		return ErrWrongPacket
	}

	if v4wire.Expired(req.Expiration) {
		return ErrExpired
	}

	if !h.checkBond(fromID, ctx.Address().(*net.UDPAddr).IP) {
		return ErrUnknownNode
	}

	p.Set(interfaces.ExpirationKey, time.Unix(int64(req.Expiration), 0)) //set the exp for rest of stack
	return nil
}

func (h *handler) HandleFindNodeRequest(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, hash []byte) error {
	return h.makeFindNodesResponse(ctx, p.GetSource().(*v4wire.Findnode), ctx.Address().(*net.UDPAddr), fromID)
}

// findNodesResponse sends a neighbour packet to the given node up to K ngbr.
// In case of buffer overflowing, it splits the response into multiple ones.
func (h *handler) makeFindNodesResponse(ctx interfaces.Context, req *v4wire.Findnode, fromIp *net.UDPAddr, fromId enode.ID) error {

	// Determine closest nodes.
	target := enode.ID(crypto.Keccak256Hash(req.Target[:]))
	closest := h.tab.FindnodeByID(target, h.NeighbourAmount, true)

	// Send neighbors in chunks with at most maxNeighbors per packet
	// to stay below the packet size limit.
	raw_packet := &v4wire.Neighbors{Expiration: req.Expiration}
	var sent bool
	for _, n := range closest.Entries() {
		err := netutil.CheckRelayIP(fromIp.IP, n.IP())
		if err != nil {
			continue
		}

		raw_packet.Nodes = append(raw_packet.Nodes, nodeToRPC(n))
		if len(raw_packet.Nodes) == v4wire.MaxNeighbors {
			if err := h.sendAndForget(ctx, fromId, fromIp, raw_packet); err != nil {
				return err
			}
			raw_packet.Nodes = raw_packet.Nodes[:0]
			sent = true
		}
	}

	if len(raw_packet.Nodes) > 0 || !sent {
		return h.sendAndForget(ctx, fromId, fromIp, raw_packet)
	}
	return nil
}

func (h *handler) VerifyFindNodeResponse(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, fromKey v4wire.Pubkey) error {
	req := p.GetSource().(*v4wire.Neighbors)
	if v4wire.Expired(req.Expiration) {
		return ErrExpired
	}

	return nil
}

func (h *handler) HandleFindNodeResponse(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, hash []byte) error {
	req := p.GetSource().(*v4wire.Neighbors)

	addr := ctx.Address()

	target := targetablePacket{
		WirePacket: p,
		exp:        int64(req.Expiration),
		id:         fromID,
		ip:         addr.(*net.UDPAddr).IP,
	}

	if err := ctx.Match(target); err != nil {
		return ErrUnsolicitedReply
	}
	return nil
}

// makes find node request over the network
// awaits for up to K neighbours.
// returns error if either the recv nodes are invalid or if none have been recv.
func (h *handler) makeFindNodeRequest(ctx interfaces.Context, id enode.ID, encKey common.EncPubkey) ([]*enode.Node, error) {
	address := ctx.Address().(*net.UDPAddr)
	request, err := h.createFindNodesRequest(id, address, v4wire.Pubkey(encKey))
	if err != nil {
		return nil, err
	}

	// active until enough nodes have been received.
	nodes := make([]*enode.Node, 0, BucketSize)
	nreceived := 0
	request.verify = func(p interfaces.Response) (bool, bool) {
		reply := p.GetSource().(*v4wire.Neighbors)
		for _, rn := range reply.Nodes {
			nreceived++
			node, err := h.nodeFromRPC(address, rn)
			if err != nil {
				continue
			}

			nodes = append(nodes, &node.Node)
		}
		return true, nreceived >= BucketSize
	}

	err = ctx.SendRequest(request, true, address)

	if err != nil && nreceived != 0 {
		err = nil
	}

	return nodes, err
}

// creates FindNode request that can be sent over the network to specified target
func (h *handler) createFindNodesRequest(toid enode.ID, toaddr *net.UDPAddr, targetKey v4wire.Pubkey) (request, error) {
	req := &v4wire.Findnode{
		Target:     targetKey,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}

	data, _, err := v4wire.Encode(h.holder.GetKey(), req)
	if err != nil {
		return request{}, err
	}

	targetable := targetablePacket{
		WirePacket: interfaces.NewWirePacket(req),
		exp:        int64(req.Expiration),
		id:         toid,
		ip:         toaddr.IP,
	}

	request := request{
		targetablePacket: targetable,
		marshalled:       data,
		expectedType:     v4wire.NeighborsPacket,
	}

	return request, err
}

func (t *handler) nodeFromRPC(sender *net.UDPAddr, rn v4wire.Node) (*node, error) {
	if rn.UDP <= 1024 {
		return nil, ErrLowPort
	}
	if err := netutil.CheckRelayIP(sender.IP, rn.IP); err != nil {
		return nil, err
	}

	key, err := v4wire.DecodePubkey(crypto.S256(), rn.ID)
	if err != nil {
		return nil, err
	}
	n := wrapNode(enode.NewV4(key, rn.IP, int(rn.TCP), int(rn.UDP)))
	err = n.ValidateComplete()
	return n, err
}

func nodeToRPC(n *common.Node) v4wire.Node {
	var key ecdsa.PublicKey
	var ekey v4wire.Pubkey
	if err := n.Load((*enode.Secp256k1)(&key)); err == nil {
		ekey = v4wire.EncodePubkey(&key)
	}
	return v4wire.Node{ID: ekey, IP: n.IP(), UDP: uint16(n.UDP()), TCP: uint16(n.TCP())}
}
