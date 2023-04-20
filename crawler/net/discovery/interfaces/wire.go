package interfaces

import (
	"net"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type WireMetadata interface {
	Ip() net.IP
	Id() enode.ID
	Kind() uint8
}

type Packet interface {
	Name() string // Name returns a string corresponding to the message type.
	Kind() byte   // Kind returns the message type.
}

// reserved ids
const (
	ExpirationKey uint8 = 1
	IdKey         uint8 = 2

	SubProtcolOffset uint8 = 10 //up to this everything is reserved
)

type WirePacket interface {
	GetSource() Packet
	Set(key uint8, val any)
	Get(key uint8) (any, bool)
}

func NewWirePacket(src Packet) WirePacket {
	return &wirePacket{src: src, data: map[uint8]any{}}
}

type wirePacket struct {
	src  Packet
	data map[uint8]any //data exchanged through different protocol stages
}

func (w wirePacket) GetSource() Packet       { return w.src }
func (w *wirePacket) Set(key uint8, val any) { w.data[key] = val }
func (w wirePacket) Get(key uint8) (any, bool) {
	d, ok := w.data[key]
	return d, ok
}

type Context interface {
	SendRequest(request Request, await bool, target net.Addr) error
	Match(Response) error
	Address() net.Addr
}

type Pubkey [64]byte

// ID returns the node ID corresponding to the public key.
func (e Pubkey) ID() enode.ID {
	return enode.ID(crypto.Keccak256Hash(e[:]))
}
