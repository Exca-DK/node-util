package base

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Exca-DK/node-util/crawler/events"
)

func NewPool(target int) *Pool {
	return &Pool{
		workers:     make([]Worker, 0),
		target:      uint32(target),
		workersIn:   make(chan Worker),
		done:        make(chan struct{}),
		workersFeed: *events.NewFeed[JobStats](),
	}
}

type Pool struct {
	workers []Worker

	target          uint32
	activeWorkers   uint32
	finishedWorkers uint32
	workersFeed     events.Feed[JobStats]

	workersIn chan Worker

	done chan struct{}

	mu sync.Mutex
}

func (p *Pool) Start(ctx context.Context) {
	go p.start(ctx)
	go p.monitor(ctx)
}

func (p *Pool) start(ctx context.Context) {
	// spawn allocated workers
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, worker := range p.workers {
		p.spawnWorker(ctx, worker)
	}
	p.workers = p.workers[:0]
}

func (p *Pool) spawnWorker(ctx context.Context, worker Worker) {
	atomic.AddUint32(&p.activeWorkers, 1)
	go func() {
		result := worker.Run(ctx)
		atomic.AddUint32(&p.activeWorkers, ^uint32(0))
		atomic.AddUint32(&p.finishedWorkers, 1)
		p.workersFeed.Send(result)
	}()
}

func (p *Pool) AddWorker(w Worker) {
	select {
	case <-p.done:
	case p.workersIn <- w:
	}
}

func (p *Pool) popWorker() Worker {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.workers) == 0 {
		return emptyWorker()
	}

	var x Worker
	x, p.workers = p.workers[0], p.workers[1:]

	return x
}

func (p *Pool) addWorker(w Worker) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.workers = append(p.workers, w)
}

func (p *Pool) monitor(ctx context.Context) {
	workerTicker := time.NewTicker(1 * time.Second)
	for {

		active := atomic.LoadUint32(&p.activeWorkers)
		if active < p.target && p.PendingWorkers() > 0 {
			p.spawnWorker(ctx, p.popWorker())
		}

		select {
		case <-ctx.Done():
			close(p.done)
			return
		case w := <-p.workersIn:
			p.addWorker(w)
		case <-workerTicker.C:
			for p.PendingWorkers() != 0 {
				active := atomic.LoadUint32(&p.activeWorkers)
				if active == p.target {
					break
				}
				p.spawnWorker(ctx, p.popWorker())
			}
		}
	}
}

func (p *Pool) RunningWorkers() int {
	return int(atomic.LoadUint32(&p.activeWorkers))
}

func (p *Pool) FinishedWorkers() int {
	return int(atomic.LoadUint32(&p.finishedWorkers))
}

func (p *Pool) PendingWorkers() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.workers)
}

func (p *Pool) SubscribeJobFeed(ch chan JobStats) events.Subscription {
	return p.workersFeed.Subscribe(ch)
}
