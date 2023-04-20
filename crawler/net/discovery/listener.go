package discovery

import (
	"errors"
	"io"
	"net"
	"sync/atomic"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/ethereum/go-ethereum/p2p/netutil"
)

const (
	maxPacketSize = 1280
)

// UDPConn is a network connection on which discovery can operate.
type UDPConn interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error)
	Close() error
	LocalAddr() net.Addr
}

type Listener interface {
	Stop()
}

func NewListener(conn UDPConn, log log.Logger, acceptor Acceptor) *listener {
	l := &listener{conn: conn, log: log, acceptor: acceptor}
	go l.loop()
	return l
}

// Listener accepts network packets and redirects them over to provided acceptor
type listener struct {
	log log.Logger

	acceptor Acceptor

	okayPackets    atomic.Uint64
	invalidPackets atomic.Uint64

	conn UDPConn
}

func (l *listener) loop() {
	buf := make([]byte, maxPacketSize)
	for {
		nbytes, from, err := l.conn.ReadFromUDP(buf)
		if netutil.IsTemporaryError(err) {
			// Ignore temporary read errors.
			l.log.Debug("Temporary UDP read error", log.NewErrorField(err))
			continue
		} else if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			l.log.Warn("UDP read error", log.NewErrorField(err))
			return
		}

		if l.acceptor.Accept(buf[:nbytes], from) != nil {
			l.invalidPackets.Add(1)
		} else {
			l.okayPackets.Add(1)
		}
	}
}

func (l *listener) Stop() {
	l.conn.Close()
}
