package interfaces

import (
	"net"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

type SeenNode interface {
	GetAddedAt() time.Time
	GetCount() int
	GetNode() *enode.Node
}

type DiscStore interface {
	StoreNode(node *enode.Node)
	GetNodes(amount int) []SeenNode
	GetLastVerifiedActivity(enode.ID, net.IP) time.Time
	GetFails(enode.ID, net.IP) int
}
