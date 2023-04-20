package interfaces

import (
	"net"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

type PacketMetadata interface {
	GetExpiration() time.Time
	Id() enode.ID
	Ip() net.IP
}

type Request interface {
	Bytes() []byte
	Verify(Response) (bool, bool)
	GetWirePacket() WirePacket
	ExpectedKind() byte
	PacketMetadata
}

type Response interface {
	GetSource() Packet
	GetExpiration() time.Time
	PacketMetadata
}

type AwaitableRequest interface {
	Await()
	Error() error
	Request() Request
}
