package v4

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/netutil"
)

func (h *handler) verifyENRRequest(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, fromKey v4wire.Pubkey) error {
	if !h.Config.ResponseEnr {
		return ErrHandlerDisabled
	}
	req := p.GetSource().(*v4wire.ENRRequest)
	if v4wire.Expired(req.Expiration) {
		return ErrExpired
	}

	if !h.checkBond(fromID, ctx.Address().(*net.UDPAddr).IP) {
		// edge case. we might went offline for short amount
		// the node we previously had estabilished conn still thinks we are up
		// and the node db might be temporary or might not include that info due to low refresh rate
		// so try to estabilish the connection
		// return unknown node anyway. next rpc should be okay after this one
		h.markFault(fromID)
		if !h.isFaulty(fromID) {
			h.makePingRequest(ctx, fromID, func() {
				h.clearFault(fromID)
			})
		}

		return ErrUnknownNode
	}
	return nil
}

func (t *handler) handleENRRequest(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, mac []byte) error {
	return t.sendAndForget(ctx, fromID, ctx.Address().(*net.UDPAddr), &v4wire.ENRResponse{
		ReplyTok: mac,
		Record:   *t.holder.Record(),
	})
}

func (h *handler) verifyEnrResponse(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, fromKey v4wire.Pubkey) error {
	target := targetablePacket{
		WirePacket: p,
		exp:        time.Now().Add(10 * time.Second).Unix(),
		id:         fromID,
		ip:         ctx.Address().(*net.UDPAddr).IP,
	}

	if err := ctx.Match(target); err != nil {
		return ErrUnsolicitedReply
	}
	return nil
}

// RequestENR sends enrRequest to the given node and waits for a response.
func (h *handler) makeEnrRequest(ctx interfaces.Context, n *enode.Node) (*enode.Node, error) {
	addr := ctx.Address().(*net.UDPAddr)
	req := &v4wire.ENRRequest{
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}
	raw, hash, err := v4wire.Encode(h.holder.GetKey(), req)
	if err != nil {
		return nil, err
	}

	var response *v4wire.ENRResponse
	verifyFunc := func(r interfaces.Response) (bool, bool) {
		converted := r.GetSource().(*v4wire.ENRResponse)
		matched := bytes.Equal(converted.ReplyTok, hash)
		if matched {
			response = converted
		}
		return matched, matched
	}

	target := targetablePacket{
		WirePacket: interfaces.NewWirePacket(req),
		exp:        int64(req.Expiration),
		id:         n.ID(),
		ip:         addr.IP,
	}

	err = ctx.SendRequest(request{
		targetablePacket: target,
		marshalled:       raw,
		verify:           verifyFunc,
		expectedType:     v4wire.ENRResponsePacket,
	}, true, addr)

	if err != nil {
		return nil, err
	}

	// Verify the response record.
	respN, err := enode.New(enode.ValidSchemes, &response.Record)
	if err != nil {
		return nil, err
	}
	if respN.ID() != n.ID() {
		return nil, fmt.Errorf("invalid ID in response record")
	}
	if respN.Seq() < n.Seq() {
		return n, nil // response record is older
	}
	if err := netutil.CheckRelayIP(addr.IP, respN.IP()); err != nil {
		return nil, fmt.Errorf("invalid IP in response record: %v", err)
	}
	return respN, nil
}
