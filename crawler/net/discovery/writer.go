package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"golang.org/x/time/rate"
)

type WriteBuilder interface {
	ThrottleIp(timeframe time.Duration, amount int) WriteBuilder
	ThrottlePackets(timeframe time.Duration, amount int) WriteBuilder
	Build() Writer
}

// handles request writes and logs additional info with metrics
// optionally can also throttle each unique packet type
type Writer interface {
	write(toaddr *net.UDPAddr, request interfaces.Request) error
}

func NewWriter(conn UDPConn, log log.Logger) Writer {
	return newWriter(conn, log)
}

// Builds writer with optional IP and PacketType throtting mechanism
func BuildWriter(conn UDPConn, log log.Logger) WriteBuilder {
	return newWriter(conn, log)
}

func newWriter(conn UDPConn, log log.Logger) *writer {
	return &writer{conn: conn, log: log, packetThrotter: throtter[byte]{}, ipThrotter: throtter[[4]byte]{}}
}

type throtter[T comparable] struct {
	amount         int
	interval       rate.Limit
	cache          sync.Map
	shouldThrottle bool
}

func (r *throtter[T]) wait(p T) {
	if !r.shouldThrottle {
		return
	}

	limiter, _ := r.cache.LoadOrStore(p, rate.NewLimiter(r.interval, r.amount))
	limiter.(*rate.Limiter).Wait(context.Background())
}

// handles request writes and logs additional info with metrics
// additionaly can dynamically throttle outgoing packets
type writer struct {
	log  log.Logger
	conn UDPConn

	packetThrotter throtter[byte]
	ipThrotter     throtter[[4]byte]
}

func (r *writer) ThrottleIp(timeframe time.Duration, amount int) WriteBuilder {
	r.ipThrotter.interval = rate.Every(timeframe)
	r.ipThrotter.amount = amount
	r.ipThrotter.shouldThrottle = true

	return r
}

func (r *writer) ThrottlePackets(timeframe time.Duration, amount int) WriteBuilder {
	r.packetThrotter.interval = rate.Every(timeframe)
	r.packetThrotter.amount = amount
	r.packetThrotter.shouldThrottle = true

	return r
}

func (r *writer) Build() Writer {
	return r
}

func (r *writer) write(toaddr *net.UDPAddr, request interfaces.Request) error {
	//throttle packets
	ts := time.Now()
	info := request.GetWirePacket().GetSource()
	r.packetThrotter.wait(info.Kind())
	if ip := toaddr.IP.To4(); ip != nil {
		r.ipThrotter.wait([4]byte(ip))
	}

	r.log.Trace(fmt.Sprintf("%v>>", info.Name()),
		log.NewStringField("name", info.Name()),
		log.NewStringField("throttledTime",
			time.Since(ts).String()),
		log.NewStringField("addr", toaddr.String()))
	_, err := r.conn.WriteToUDP(request.Bytes(), toaddr)
	return err
}

type testWriter struct {
	remote *net.UDPAddr
	w      Writer
}

func (t *testWriter) write(toaddr *net.UDPAddr, request interfaces.Request) error {
	if !toaddr.IP.Equal(t.remote.IP) {
		return errors.New("invalid ip")
	}
	return t.w.write(toaddr, request)
}
