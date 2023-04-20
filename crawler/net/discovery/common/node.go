package common

import (
	"net"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// Node represents a host on the network.
// The fields of Node may not be modified.
type Node struct {
	enode.Node
	addedAt        uint64 // time when the node was first seen
	livenessChecks uint64 // how often liveness was checked
}

func (n *Node) SetTimestamp(ts time.Time) {
	n.addedAt = uint64(ts.Unix())
}

func (n *Node) Bump() {
	n.livenessChecks++
}

func (n *Node) BumpX(i int) {
	n.livenessChecks += uint64(i)
}

func (n *Node) Unwrap() *enode.Node {
	return &n.Node
}

func (n *Node) Wrap(node *enode.Node) {
	n.Node = *node
}

func WrapNode(n *enode.Node) *Node {
	return &Node{Node: *n}
}

func WrapNodes(ns []*enode.Node) []*Node {
	result := make([]*Node, len(ns))
	for i, n := range ns {
		node := &Node{}
		node.Wrap(n)
		result[i] = node
	}
	return result
}

func UnwrapNode(n *Node) *enode.Node {
	return &n.Node
}

func UnwrapNodes(ns []*Node) []*enode.Node {
	result := make([]*enode.Node, len(ns))
	for i, n := range ns {
		result[i] = n.Unwrap()
	}
	return result
}

func (n *Node) Addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.IP(), Port: n.UDP()}
}

func (n *Node) String() string {
	return n.Node.String()
}

func (n Node) GetAddedAt() time.Time { return time.Unix(int64(n.addedAt), 0) }
func (n *Node) GetCount() int        { return int(n.livenessChecks) }
func (n *Node) GetNode() *enode.Node { return &n.Node }

// NodesByDistance is a list of nodes, ordered by distance to target.
type NodesByDistance struct {
	entries []*Node
	Target  enode.ID
}

func (n *NodesByDistance) Entries() []*Node {
	return n.entries
}

func (n *NodesByDistance) Enodes() []*enode.Node {
	nodes := make([]*enode.Node, len(n.entries))
	for i, entry := range n.entries {
		nodes[i] = entry.Unwrap()
	}
	return nodes
}

// push adds the given node to the list, keeping the total size below maxElems.
func (h *NodesByDistance) Push(n *Node, maxElems int) {
	ix := sort.Search(len(h.entries), func(i int) bool {
		return enode.DistCmp(h.Target, h.entries[i].ID(), n.ID()) > 0
	})

	end := len(h.entries)
	if len(h.entries) < maxElems {
		h.entries = append(h.entries, n)
	}
	if ix < end {
		// Slide existing entries down to make room.
		// This will overwrite the entry we just appended.
		copy(h.entries[ix+1:], h.entries[ix:])
		h.entries[ix] = n
	}
}
