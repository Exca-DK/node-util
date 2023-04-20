package interfaces

import "github.com/ethereum/go-ethereum/p2p/enode"

type Iterator interface {
	Next() bool
	Node() *enode.Node
	Close()
}
