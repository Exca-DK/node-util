package table

import (
	"net"

	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type Tab interface {
	AddSeenNode(n *common.Node)
	FindnodeByID(target enode.ID, nresults int, preferLive bool) *common.NodesByDistance
	FindFails(id enode.ID, ip net.IP) int
	UpdateFindFails(id enode.ID, ip net.IP, fails int) error
	BucketLen(id enode.ID) int
	Self() *enode.Node
}
