package discovery

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

func NewRandomIdentity(name string, ip net.IP, tcp uint16, udp uint16, log log.Logger) *identity {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(fmt.Errorf("key gen failure. error: %v", err))
	}

	i := &identity{
		name:    name,
		key:     key,
		ip:      ip,
		tcp:     tcp,
		udp:     udp,
		log:     log,
		entries: map[string]enr.Entry{},
	}
	i.dirty.Store(true)

	i.SetEntry(enr.IPv4(ip))
	i.SetEntry(enr.UDP(udp))
	i.SetEntry(enr.TCP(tcp))

	i.sign()
	return i
}

type identity struct {
	key      *ecdsa.PrivateKey
	ip       net.IP
	tcp, udp uint16
	log      log.Logger

	name string

	dirty atomic.Bool
	cur   atomic.Value

	mu      sync.Mutex
	entries map[string]enr.Entry
	seq     uint64
}

func (i *identity) FakeName() string {
	if i.name == "" {
		return "geth/linux-amd64/go1.18.1"
	}
	return i.name
}
func (i *identity) GetKey() *ecdsa.PrivateKey { return i.key }
func (i *identity) GetIp() net.IP             { return i.ip }
func (i *identity) GetTcp() uint16            { return i.udp }
func (i *identity) GetUdp() uint16            { return i.tcp }

// Seq returns the current sequence number of the local node record.
func (ln *identity) GetSequence() uint64 {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	return ln.seq
}

func (i *identity) Record() *enr.Record { return i.GetNode().Record() }
func (i *identity) GetNode() *enode.Node {
	if !i.dirty.Load() {
		return i.cur.Load().(*enode.Node)
	}
	return i.sign()
}

func (i *identity) SetEntry(e enr.Entry) {
	i.mu.Lock()
	defer i.mu.Unlock()

	val, exists := i.entries[e.ENRKey()]
	if !exists || !reflect.DeepEqual(val, e) {
		i.entries[e.ENRKey()] = e
		i.dirty.Store(true)
	}
}

func (i *identity) sign() *enode.Node {
	i.mu.Lock()
	defer i.mu.Unlock()
	//something else already did it. safe to back
	if !i.dirty.Load() {
		return i.cur.Load().(*enode.Node)
	}

	var r enr.Record
	for _, e := range i.entries {
		r.Set(e)
	}
	i.seq++
	r.SetSeq(i.seq)

	if err := enode.SignV4(&r, i.key); err != nil {
		panic(fmt.Errorf("enode: can't sign record: %v", err))
	}

	n, err := enode.New(enode.ValidSchemes, &r)
	if err != nil {
		panic(fmt.Errorf("enode: can't verify local record: %v", err))
	}

	i.cur.Store(n)
	i.dirty.Store(false)
	return n
}
