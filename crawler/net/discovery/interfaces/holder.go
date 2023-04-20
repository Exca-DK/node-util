package interfaces

import (
	"crypto/ecdsa"
	"net"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

type IdentityHolder interface {
	FakeName() string
	GetKey() *ecdsa.PrivateKey
	GetIp() net.IP
	GetTcp() uint16
	GetUdp() uint16
	Record() *enr.Record
	GetNode() *enode.Node
	GetSequence() uint64
}
