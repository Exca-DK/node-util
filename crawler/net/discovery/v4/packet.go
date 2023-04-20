package v4

import (
	"net"
	"time"

	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type targetablePacket struct {
	interfaces.WirePacket
	exp int64
	id  enode.ID
	ip  net.IP
}

func (r targetablePacket) GetExpiration() time.Time {
	return time.Unix(r.exp, 0)
}

func (r targetablePacket) Id() enode.ID {
	return r.id
}

func (r targetablePacket) Ip() net.IP {
	return r.ip
}

type request struct {
	targetablePacket
	marshalled   []byte
	verify       func(interfaces.Response) (bool, bool)
	expectedType byte
}

func (r request) Bytes() []byte {
	return r.marshalled
}

func (r request) Verify(res interfaces.Response) (bool, bool) {
	return r.verify(res)
}

func (r request) GetWirePacket() interfaces.WirePacket {
	return r
}

func (r request) ExpectedKind() byte {
	return r.expectedType
}
