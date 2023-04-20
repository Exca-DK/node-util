package discovery

import (
	"net"

	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// ctx is context that holds three basic functionalities for sub handlers
// It includes Address which is either target address or recv address depending on context
// And has functionalities to send awaitable request and match packets to them
type ctx struct {
	send   func(interfaces.Request, bool, net.Addr) error
	match  func(interfaces.Response) error
	listen func(func() bool, func(interfaces.Response) bool, enode.ID, net.IP) error
	ip     net.Addr
}

// Send request to specified address. If await flag is set, the response is automatically awaited.
// Will return error in two cases:
// 1.   none of the response match the provided verification function of request
// 2.   awaiter is unavaiable
func (ctx ctx) SendRequest(r interfaces.Request, await bool, addr net.Addr) error {
	return ctx.send(r, await, addr)
}

// Tries to match response over to request
func (ctx ctx) Match(r interfaces.Response) error { return ctx.match(r) }

// Depending on the context it can return either of two:
// 1. returns address of packet sender
// 2. returns address of packet to send
func (ctx ctx) Address() net.Addr { return ctx.ip }

// Optimisticly await for packet from node with specified ip and id
func (ctx ctx) AwaitOptimistic(done func() bool, ip net.IP, id enode.ID, finished func(interfaces.Response) bool) bool {
	return ctx.listen(done, finished, id, ip) == nil
}
