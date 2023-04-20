package v4

import (
	"bytes"
	"crypto/ecdsa"
	"net"
	"time"

	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
)

func (h *handler) VerifyPingRequest(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, fromKey v4wire.Pubkey) error {
	addr := ctx.Address()

	return h.verifyPing(addr.(*net.UDPAddr), fromKey, p)
}

func (h *handler) verifyPing(from *net.UDPAddr, fromkey v4wire.Pubkey, p interfaces.WirePacket) error {
	req, ok := p.GetSource().(*v4wire.Ping)
	if !ok {
		return ErrWrongPacket
	}

	if v4wire.Expired(req.Expiration) {
		return ErrExpired
	}

	sender, err := v4wire.DecodePubkey(crypto.S256(), v4wire.Pubkey(fromkey))
	if err != nil {
		return err
	}

	p.Set(senderKey, sender)

	h.store.UpdateLastPingReceived(fromkey.ID(), from.IP, time.Now())
	return nil
}

// Responds to the ping with pong
// Might also execute ping request to the sender if it's considered not seen within timeframe.
func (h *handler) HandlePingRequest(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, hash []byte) error {
	ping, ok := p.GetSource().(*v4wire.Ping)
	if !ok {
		return ErrWrongPacket
	}

	addr := ctx.Address().(*net.UDPAddr)

	pong := &v4wire.Pong{
		To:         v4wire.NewEndpoint(addr, ping.From.TCP),
		ReplyTok:   hash,
		Expiration: ping.Expiration,
		ENRSeq:     h.Self().Seq(),
	}

	raw, _, err := v4wire.Encode(h.holder.GetKey(), pong)
	if err != nil {
		return err
	}

	req := request{targetablePacket: targetablePacket{WirePacket: interfaces.NewWirePacket(pong), id: fromID}, marshalled: raw}
	if err := ctx.SendRequest(req, false, addr); err != nil {
		return err
	}

	sender, _ := p.Get(senderKey)
	n := common.WrapNode(enode.NewV4(sender.(*ecdsa.PublicKey), addr.IP, int(ping.From.TCP), addr.Port))
	if time.Since(h.store.LastPongReceived(n.ID(), addr.IP)) > bondExpiration {
		_, err := h.makePingRequest(ctx, fromID, func() {
			h.tab.AddSeenNode(n)
		})
		if err != nil {
			return err
		}
	} else {
		h.tab.AddSeenNode(n)
	}

	return nil
}

func (h *handler) VerifyPongResponse(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, fromKey v4wire.Pubkey) error {
	addr := ctx.Address()
	return h.verifyPong(addr.(*net.UDPAddr), fromKey, p)
}

func (h *handler) verifyPong(from *net.UDPAddr, key v4wire.Pubkey, p interfaces.WirePacket) error {
	req, ok := p.GetSource().(*v4wire.Pong)
	if !ok {
		return ErrWrongPacket
	}

	if v4wire.Expired(req.Expiration) {
		return ErrExpired
	}

	h.store.UpdateLastPongReceived(key.ID(), from.IP, time.Now())
	return nil
}

func (h *handler) HandlePongResponse(ctx interfaces.Context, p interfaces.WirePacket, fromID enode.ID, hash []byte) error {
	pong, ok := p.GetSource().(*v4wire.Pong)
	if !ok {
		return ErrUnsolicitedReply
	}
	addr := ctx.Address()

	target := targetablePacket{
		WirePacket: p,
		exp:        int64(pong.Expiration),
		id:         fromID,
		ip:         addr.(*net.UDPAddr).IP,
	}

	return ctx.Match(target)
}

// ping sends a ping message to the given node and waits for a reply.
// Ctx should include target address
func (h *handler) makePingRequest(ctx interfaces.Context, id enode.ID, callback func()) (uint64, error) {
	request, err := h.generatePing(id, ctx.Address().(*net.UDPAddr), callback)
	if err != nil {
		return 0, err
	}

	err = ctx.SendRequest(request, true, ctx.Address())
	if err != nil {
		return 0, err
	}

	packet := request.GetWirePacket()
	val, _ := packet.Get(enrSeqKey)
	return val.(uint64), nil
}

// sendPing sends a ping message to the given node and invokes the callback
// when the reply arrives.
func (h *handler) generatePing(toid enode.ID, toaddr *net.UDPAddr, callback func()) (interfaces.Request, error) {
	req := &v4wire.Ping{
		Version:    4,
		From:       v4wire.Endpoint{IP: h.holder.GetIp(), UDP: h.holder.GetUdp(), TCP: h.holder.GetTcp()},
		To:         v4wire.NewEndpoint(toaddr, 0),
		Expiration: uint64(time.Now().Add(expiration).Unix()),
		ENRSeq:     h.holder.GetSequence(),
		Rest:       []rlp.RawValue{},
	}

	data, hash, err := v4wire.Encode(h.holder.GetKey(), req)
	if err != nil {
		return nil, err
	}

	wire := interfaces.NewWirePacket(req)
	verifierFunc := func(p interfaces.Response) (bool, bool) {
		response_packet := p.GetSource().(*v4wire.Pong)
		matched := bytes.Equal(response_packet.ReplyTok, hash)
		if !matched {
			return false, false
		}
		if callback != nil {
			callback()
		}

		wire.Set(enrSeqKey, response_packet.ENRSeq)
		return true, true
	}

	target := targetablePacket{
		WirePacket: wire,
		exp:        int64(req.Expiration),
		id:         toid,
		ip:         toaddr.IP,
	}

	return request{
		targetablePacket: target,
		marshalled:       data,
		verify:           verifierFunc,
		expectedType:     v4wire.PongPacket,
	}, nil
}
