package net

import (
	"fmt"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/Exca-DK/node-util/crawler/net/transport"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var (
	DefaultQuieries = []p2p.Cap{
		{Name: "diff", Version: 1},
		{Name: "eth", Version: 64},
		{Name: "eth", Version: 65},
		{Name: "eth", Version: 66},
		{Name: "les", Version: 2},
		{Name: "les", Version: 3},
		{Name: "les", Version: 4},
		{Name: "snap", Version: 1},
	}
)

func NewPeerQuierier(id interfaces.IdentityHolder, log log.Logger) *PeerQuierier {
	return &PeerQuierier{id: id, log: log}
}

type PeerQuierier struct {
	id interfaces.IdentityHolder

	extraQuery func(*transport.Transport) map[string]string //get extra return

	log log.Logger
	// exchanged with peers
	caps []p2p.Cap
}

type PeerInfo struct {
	Name    string
	Caps    []string
	Version int
	Extra   map[string]string
}

func (p *PeerQuierier) Query(node *enode.Node) (PeerInfo, error) {
	transport, err := transport.Dial(node, p.id.GetKey())
	if err != nil {
		return PeerInfo{}, fmt.Errorf("failed dial. peer: %v error: %v", node.String(), err)
	}
	defer transport.Close()
	if err := transport.Setup(p.id.FakeName(), p.caps); err != nil {
		return PeerInfo{}, fmt.Errorf("failed transport setup peer: %v error: %v", node.String(), err)
	}

	info := PeerInfo{Name: transport.Name, Version: int(transport.Version)}
	caps := transport.Caps
	info.Caps = make([]string, len(caps))
	for i, cap := range caps {
		info.Caps[i] = cap.String()
	}

	if p.extraQuery != nil {
		info.Extra = p.extraQuery(transport)
	}

	return info, nil
}

func (p *PeerQuierier) AddCap(name string, version uint) {
	p.caps = append(p.caps, p2p.Cap{Name: name, Version: version})
}
