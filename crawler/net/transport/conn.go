package transport

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/rlpx"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
)

// dial attempts to dial the given node and perform a handshake with hello,
func Dial(n *enode.Node, key *ecdsa.PrivateKey) (*Transport, error) {
	fd, err := net.Dial("tcp", fmt.Sprintf("%v:%d", n.IP(), n.TCP()))
	if err != nil {
		return nil, err
	}

	rlpconn := rlpx.NewConn(fd, n.Pubkey())
	if err = rlpconn.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return nil, errors.Wrap(err, "cannot set conn deadline")
	}
	_, err = rlpconn.Handshake(key)
	if err != nil {
		return nil, err
	}

	conn := &Conn{conn: rlpconn, ourKey: key}

	transport := &Transport{Conn: conn}
	return transport, nil
}

// Conn represents an individual connection with a peer
type Conn struct {
	conn   *rlpx.Conn
	ourKey *ecdsa.PrivateKey

	wmu, rmu sync.Mutex
}

func (c *Conn) Read() (Msg, error) {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	code, data, _, err := c.conn.Read()
	var msg Msg
	if err == nil {
		msg.ReceivedAt = time.Now().UnixMilli()
		msg.Payload = bytes.NewReader(copyb(data))
		msg.code = code
		msg.Size = uint32(len(data))
	}
	return msg, err
}

func (c *Conn) Write(msg Message) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	payload, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(uint64(msg.Code()), payload)
	return err
}

func (t *Conn) Close() {
	t.wmu.Lock()
	defer t.wmu.Unlock()
	t.conn.Close()
}

func copyb(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}
