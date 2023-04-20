package discovery

import (
	"context"
	"crypto/ecdsa"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery/common"
	v4 "github.com/Exca-DK/node-util/crawler/net/discovery/v4"
	"github.com/Exca-DK/node-util/crawler/net/sims"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func newkey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key
}

func TestUDPv4_ping(t *testing.T) {
	t.Parallel()

	p1, p2 := sims.NewUdpPipe()
	handler1 := newTestHandler(p1, Config{})
	handler2 := newTestHandler(p2, Config{})

	l1 := NewListener(p1, log.NewLogger(), handleracceptor{handler1})
	l2 := NewListener(p2, log.NewLogger(), handleracceptor{handler2})
	defer l2.Stop()
	defer l1.Stop()

	_, err := handler1.Ping(handler2.Self())
	if err != nil {
		t.Fatal(err)
	}
}

func TestUDPv4_pingTimeout(t *testing.T) {
	t.Parallel()

	p1, p2 := sims.NewUdpPipe()
	handler1 := newTestHandler(p1, Config{})
	handler2 := newTestHandler(p2, Config{})

	l1 := NewListener(p1, log.NewLogger(), handleracceptor{handler1})
	l2 := NewListener(p2, log.NewLogger(), dummyAcceptor{})
	defer l2.Stop()
	defer l1.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	errch := make(chan error, 1)
	go func() {
		_, err := handler1.Ping(handler2.Self())
		errch <- err
	}()

	select {
	case <-ctx.Done():
	case err := <-errch:
		if err == nil {
			t.Fatal("expected time-out")
		}
	}

}

func TestUDPv4_findNode(t *testing.T) {
	t.Parallel()
	log.SetLoggingLvl(log.TRACE)
	p1, p2 := sims.NewUdpPipe()
	bootnodes := []*enode.Node{}
	for i := 0; i < 8; i++ {
		c1, c2 := sims.NewUdpPipe()
		node1 := enode.NewV4(&newkey().PublicKey, c1.LocalIp(), c1.LocalPort(), c1.LocalPort())
		node2 := enode.NewV4(&newkey().PublicKey, c2.LocalIp(), c2.LocalPort(), c2.LocalPort())
		bootnodes = append(bootnodes, node1, node2)
	}

	handler1 := newTestHandler(p1, Config{})
	handler1.Run()
	handler2 := newTestHandler(p2, Config{Bootnodes: bootnodes})
	handler2.Run()
	l1 := NewListener(p1, log.NewLogger(), NewPoolAcceptor(handler1, 4))
	l2 := NewListener(p2, log.NewLogger(), NewPoolAcceptor(handler2, 4))
	defer l2.Stop()
	defer l1.Stop()

	nodes, err := handler1.AskForRandomNodes(handler2.Self())
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 16 {
		t.Fatalf("missing nodes. got: %v, want: %v\n", len(nodes), len(bootnodes))
	}
}

func TestUDPv4_findNodeNotFull(t *testing.T) {
	t.Parallel()
	p1, p2 := sims.NewUdpPipe()
	bootnodes := []*enode.Node{}
	for i := 0; i < 2; i++ {
		c1, c2 := sims.NewUdpPipe()
		node1 := enode.NewV4(&newkey().PublicKey, c1.LocalIp(), c1.LocalPort(), c1.LocalPort())
		node2 := enode.NewV4(&newkey().PublicKey, c2.LocalIp(), c2.LocalPort(), c2.LocalPort())
		bootnodes = append(bootnodes, node1, node2)
	}

	handler1 := newTestHandler(p1, Config{})
	handler2 := newTestHandler(p2, Config{Bootnodes: bootnodes})

	l1 := NewListener(p1, log.NewLogger(), NewPoolAcceptor(handler1, 4))
	l2 := NewListener(p2, log.NewLogger(), NewPoolAcceptor(handler2, 4))
	defer l2.Stop()
	defer l1.Stop()

	nodes, err := handler1.AskForRandomNodes(handler2.Self())
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) < 4 {
		t.Fatalf("missing nodes. got: %v, want: %v\n", len(nodes), len(bootnodes))
	}
}

func TestUDPv4_enr(t *testing.T) {
	t.Parallel()
	p1, p2 := sims.NewUdpPipe()

	handler1 := newTestHandler(p1, Config{})
	handler2 := newTestHandler(p2, Config{})
	l1 := NewListener(p1, log.NewLogger(), NewPoolAcceptor(handler1, 4))
	l2 := NewListener(p2, log.NewLogger(), NewPoolAcceptor(handler2, 4))
	defer l2.Stop()
	defer l1.Stop()
	enr, err := handler1.RequestENR(handler2.Self())
	if err != nil {
		t.Fatal(err)
	}

	target := handler2.Self()

	if !reflect.DeepEqual(enr, target) {
		t.Fatal("different enr result")
	}
}

func TestNodeAddedEvent(t *testing.T) {
	t.Parallel()
	p1, _ := sims.NewUdpPipe()

	bootnodes := []*enode.Node{}
	for i := 0; i < 8; i++ {
		c1, c2 := sims.NewUdpPipe()
		node1 := enode.NewV4(&newkey().PublicKey, c1.LocalIp(), c1.LocalPort(), c1.LocalPort())
		node2 := enode.NewV4(&newkey().PublicKey, c2.LocalIp(), c2.LocalPort(), c2.LocalPort())
		bootnodes = append(bootnodes, node1, node2)
	}
	handler := newTestHandler(p1, Config{RunTab: true})

	ch, sub := handler.SubscribeNodeEvents()
	defer sub.Unsubscribe()
	handler.Run()
	go func() {
		for _, node := range bootnodes {
			handler.table.AddSeenNode(common.WrapNode(node))
		}
	}()

	c := 0
	for {
		val := <-ch
		if val.Type == ADDED {
			c++
		}

		if c == len(bootnodes) {
			break
		}
	}
}

func newTestHandler(conn *sims.UdpConn, cfg Config) *PacketHandler {
	identity := NewRandomIdentity("", conn.LocalIp(), uint16(conn.LocalPort()), uint16(conn.LocalPort()), log.NewLogger())
	db, err := enode.OpenDB("")
	if err != nil {
		panic(err)
	}

	writer := &testWriter{w: NewWriter(conn, log.NewLogger()), remote: conn.RemoteAddr().(*net.UDPAddr)}
	h, err := NewHandler(writer, v4.NewHandler(identity, log.NewLoggerWithId(conn.LocalIp().String()), v4.Config{NeighbourAmount: 16, ResponseEnr: true, ResponseFindNode: true}), db, cfg)
	if err != nil {
		panic(err)
	}

	return h
}
