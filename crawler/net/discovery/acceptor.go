package discovery

import (
	"errors"
	"net"
	"sync"
	"time"
)

// Acceptor is an interface abstracting handling from recv network packets
type Acceptor interface {
	Accept(data []byte, from *net.UDPAddr) error
}

// accetor that directly passes packets into handler
type handleracceptor struct {
	handler *PacketHandler
}

func (a handleracceptor) Accept(data []byte, from *net.UDPAddr) error {
	return a.handler.HandlePacket(data, from)
}

// usefull for simulating no-response
type dummyAcceptor struct{}

func (a dummyAcceptor) Accept(data []byte, from *net.UDPAddr) error {
	return nil
}

// usefull for simulating latency issues by using method with eg. sleep
type functionalAcceptor struct {
	f func(data []byte, from *net.UDPAddr) error
}

func (a functionalAcceptor) Accept(data []byte, from *net.UDPAddr) error {
	return a.f(data, from)
}

type acceptorJob struct {
	data []byte
	from *net.UDPAddr
}

func NewPoolAcceptor(handler IHandler, workers int) *poolAcceptor {
	return &poolAcceptor{
		handler: handler,
		chJob:   make(chan acceptorJob),
		stop:    make(chan struct{}),
		chDone:  make(chan struct{}),
		workers: workers,
	}
}

func NewPoolAcceptorWithLimits(handler IHandler, workers int, rate_timeframe time.Duration, packetsBound int) *poolAcceptor {
	return &poolAcceptor{
		limit: true,
		limiter: limiter{
			timeDeltaTh: rate_timeframe,
			bound:       packetsBound,
			m:           make(map[[4]byte]acceptLimit),
		},
		handler: handler,
		chJob:   make(chan acceptorJob),
		stop:    make(chan struct{}),
		chDone:  make(chan struct{}),
		workers: workers,
	}
}

// accetor that passes packet into worker
type poolAcceptor struct {
	limit   bool
	limiter limiter
	handler IHandler
	once    sync.Once
	chJob   chan acceptorJob
	stop    chan struct{}
	chDone  chan struct{}
	workers int
}

func (a *poolAcceptor) start() {
	for i := 0; i < a.workers; i++ {
		go a.listen()
	}

	go func() {
		<-a.stop
		close(a.chDone)
	}()
}

func (a *poolAcceptor) listen() {
	for {
		select {
		case <-a.chDone:
			return
		case job := <-a.chJob:
			a.handler.HandlePacket(job.data, job.from)
		}
	}
}

func (a *poolAcceptor) Stop() {
	select {
	case <-a.chDone:
	case a.stop <- struct{}{}:
	}
}

func (a *poolAcceptor) send(data []byte, from *net.UDPAddr) {
	select {
	case <-a.chDone:
	case a.chJob <- acceptorJob{
		data: data,
		from: from,
	}:
	}
}

func (a *poolAcceptor) Accept(data []byte, from *net.UDPAddr) error {
	a.once.Do(a.start)
	if a.limit {
		a.limiter.update(from.IP)
		if a.limiter.isAbove(from.IP) {
			return errors.New("rate limit excedeed")
		}
	}

	buff := make([]byte, len(data))
	copy(buff, data)
	a.send(buff, from)
	return nil
}

type acceptLimit struct {
	ip   net.IP
	c    int
	last int64
}

type limiter struct {
	timeDeltaTh time.Duration
	bound       int
	m           map[[4]byte]acceptLimit
}

func (l *limiter) update(ip net.IP) {
	v, ok := l.m[[4]byte(ip.To4())]
	if !ok {
		v = acceptLimit{ip: ip}
	}
	v.c++
	if elapsed := time.Since(time.Unix(v.last, 0)); elapsed > l.timeDeltaTh {
		v.c -= int(elapsed/(l.timeDeltaTh*l.timeDeltaTh)) + 1
		if v.c < 0 {
			v.c = 0
		}
	}
	v.last = time.Now().Unix()
	l.m[[4]byte(ip.To4())] = v
}

func (l *limiter) isAbove(ip net.IP) bool {
	v := l.m[[4]byte(ip.To4())]
	return v.c > l.bound
}
