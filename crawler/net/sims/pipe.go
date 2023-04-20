package sims

import (
	"crypto/rand"
	"net"

	"github.com/ethereum/go-ethereum/p2p/netutil"
)

func NewUdpPipe() (*UdpConn, *UdpConn) {
	c1, c2 := net.Pipe()

	ip1 := make(net.IP, net.IPv4len)
	ip2 := make(net.IP, net.IPv4len)

	// find some valid random addresses
	for {
		rand.Read(ip1)
		rand.Read(ip2)
		if netutil.CheckRelayIP(ip1, ip2) != nil {
			continue
		}
		if netutil.CheckRelayIP(ip2, ip1) != nil {
			continue
		}
		break
	}

	addr1 := &net.UDPAddr{IP: ip1, Port: 5050}
	addr2 := &net.UDPAddr{IP: ip2, Port: 5050}

	return &UdpConn{rAddr: addr2, lAddr: addr1, Conn: c1}, &UdpConn{rAddr: addr1, lAddr: addr2, Conn: c2}
}

type UdpConn struct {
	rAddr *net.UDPAddr
	lAddr *net.UDPAddr
	Conn  net.Conn
}

// ReadFromUDP acts like ReadFrom but returns a UDPAddr.
func (c *UdpConn) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	return c.readFromUDP(b)
}

// readFromUDP implements ReadFromUDP.
func (c *UdpConn) readFromUDP(b []byte) (int, *net.UDPAddr, error) {
	n, err := c.Conn.Read(b)

	return n, c.rAddr, err
}

func (c *UdpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error) { return c.Conn.Write(b) }
func (c *UdpConn) Close() error                                              { return c.Conn.Close() }
func (c *UdpConn) LocalAddr() net.Addr                                       { return c.lAddr }
func (c *UdpConn) LocalIp() net.IP                                           { return c.lAddr.IP }
func (c *UdpConn) LocalPort() int                                            { return c.lAddr.Port }
func (c *UdpConn) RemoteAddr() net.Addr                                      { return c.rAddr }
func (c *UdpConn) RemotePort() int                                           { return c.rAddr.Port }
func (c *UdpConn) RemoteIp() net.IP                                          { return c.rAddr.IP }
