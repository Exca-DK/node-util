package discovery

import (
	"container/list"
	"errors"
	"net"

	"sync/atomic"
	"time"

	"github.com/Exca-DK/node-util/crawler/net/discovery/interfaces"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Exca-DK/node-util/crawler/log"
)

var (
	errRespAwaiterDone    = errors.New("response awaiter done")
	errRespDeadlinePassed = errors.New("response deadline excedeed")
	errInvalidResponse    = errors.New("response invalid")
	errAlreadyRunning     = errors.New("already running")
)

type responseListener struct {
	expired    func() bool
	sub        chan struct{}
	targetIp   net.IP
	targetId   enode.ID
	targetType uint8
}

type awaitableRequest struct {
	r       interfaces.Request
	err     error
	checker func(interfaces.Response) (bool, bool)
	ch      chan struct{}
}

func (m *awaitableRequest) mark(err error) {
	m.err = err
	close(m.ch)
}

func (m *awaitableRequest) Await() {
	<-m.ch
}

func (m *awaitableRequest) Error() error {
	return m.err
}

func (m *awaitableRequest) Request() interfaces.Request {
	return m.r
}

func NewAwaiter(logger log.Logger) *awaiter {
	a := &awaiter{
		queue:   list.New(),
		running: atomic.Bool{},
		replyCh: make(chan interfaces.Response),
		reqCh:   make(chan *awaitableRequest),
		done:    make(chan struct{}),
		logger:  logger,
	}

	a.Start()
	return a
}

type awaiter struct {
	queue *list.List

	running atomic.Bool

	replyCh chan interfaces.Response
	reqCh   chan *awaitableRequest

	done chan struct{}

	logger log.Logger
}

func (r *awaiter) Start() error {
	if !r.running.CompareAndSwap(false, true) {
		return errAlreadyRunning
	}

	go r.run()
	return nil
}

func (r *awaiter) run() {
	lower, upper := 100*time.Millisecond, 5*time.Second
	current := upper / 2
	ticker := time.NewTicker(current)

OUTER:
	for {
		select {
		case reply := <-r.replyCh:
			el := r.queue.Front()
			for el != nil {
				val := el.Value.(*awaitableRequest)
				req := val.Request()

				if req.Id() != reply.Id() || req.ExpectedKind() != reply.GetSource().Kind() || !req.Ip().Equal(reply.Ip()) {
					el = el.Next()
					continue
				}

				if time.Now().After(req.GetExpiration()) {
					r.queue.Remove(el)
					val.mark(errRespDeadlinePassed)
					break
				}

				ok, done := val.checker(reply)
				if done && ok {
					r.queue.Remove(el)
					val.mark(nil)
					r.logger.Debug("removed awaitable", log.NewStringField("reason", "matched fully"), log.NewStringField("id", req.Id().String()))
				}

				if !ok && !done {
					val.mark(errInvalidResponse)
					r.queue.Remove(el)
				}

				break
			}
		case req := <-r.reqCh:
			r.queue.PushBack(req)
		case <-ticker.C:
			count := r.checkDeadlines()
			if count == 0 {
				current *= 2
				if current > upper {
					current = upper
				}
			} else {
				current -= current / 3
				if current < lower {
					current = lower
				}
			}
			ticker.Reset(current)
		case <-r.done:
			close(r.done)
			break OUTER
		}
	}

	r.logger.Debug("awaiter closed down")
}

func (r *awaiter) checkDeadlines() int {
	now := time.Now()
	counter := 0
	q := r.queue.Front()
	if q == nil {
		return 0
	}

	for {
		el := q.Next()
		if el == nil {
			break
		}
		val := el.Value.(*awaitableRequest)
		deadline := val.r.GetExpiration()
		if now.After(deadline) || now.Equal(deadline) {
			r.queue.Remove(el)
			counter++
			val.mark(errRespDeadlinePassed)
		} else {
			break
		}
	}

	val := q.Value.(*awaitableRequest)
	deadline := val.r.GetExpiration()
	if now.After(deadline) || now.Equal(deadline) {
		r.queue.Remove(q)
		counter++
		val.mark(errRespDeadlinePassed)
	}

	return counter
}

func (r *awaiter) AddPending(request interfaces.Request, checker func(interfaces.Response) (bool, bool)) interfaces.AwaitableRequest {
	awaitable := &awaitableRequest{
		r:       request,
		checker: checker,
		ch:      make(chan struct{}),
	}
	select {
	case <-r.done:
		awaitable.mark(errRespAwaiterDone)
	case r.reqCh <- awaitable:
	}
	return awaitable
}

func (r *awaiter) MatchResponse(request interfaces.Response) error {
	select {
	case <-r.done:
		return errRespAwaiterDone
	case r.replyCh <- request:
	}
	return nil
}
