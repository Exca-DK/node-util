package runner

import (
	"context"
	"errors"
	"sync"
	"time"

	"fmt"

	"github.com/Exca-DK/node-util/crawler/log"
	"github.com/Exca-DK/node-util/crawler/metrics"
	"github.com/Exca-DK/node-util/crawler/utils"
	"github.com/Exca-DK/node-util/crawler/workers/base"
)

type NOP struct{}

func NewReader[T WorkItem](threads int, recv IteratorRecv[T], workerExit func(feed base.JobStats)) Runner[T, NOP] {
	return newRunner[T, NOP](threads, recv, nil, workerExit, false)
}

func NewReaderWriter[T WorkItem, Y WorkItem](threads int, recv IteratorRecv[T], sender IteratorSend[Y], workerExit func(feed base.JobStats)) Runner[T, Y] {
	return newRunner(threads, recv, sender, workerExit, true)
}

func newRunner[T WorkItem, Y any](threads int, recv IteratorRecv[T], sender IteratorSend[Y], workerExit func(feed base.JobStats), write bool) Runner[T, Y] {
	if workerExit == nil {
		workerExit = func(feed base.JobStats) {}
	}

	return &runner[T, Y]{
		workPool:        base.NewPool(threads),
		itRecv:          recv,
		itSend:          sender,
		sig:             make(chan struct{}),
		done:            make(chan struct{}),
		logger:          log.NewLogger(),
		write:           write,
		afterExitWorker: workerExit,
		metrics:         metrics.Enabled(),
	}
}

type runner[T WorkItem, Y any] struct {
	workPool *base.Pool

	itRecv IteratorRecv[T]
	itSend IteratorSend[Y]
	write  bool

	sig  chan struct{}
	done chan struct{}

	metrics  bool
	recorder RunnerRecorder

	logger log.Logger

	afterExitWorker func(feed base.JobStats)

	wg sync.WaitGroup
}

func (r *runner[T, Y]) Start(worker SourceWorker[T, Y]) {
	r.wg.Add(1)
	if r.metrics {
		r.recorder = newRecorder(worker.Identity(), r.write)
	} else {
		r.recorder = noOpRecorder(worker.Identity())
	}
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		r.workPool.Start(ctx)
		defer cancel()

		r.runWorkers(worker)
		<-r.sig
		close(r.done)
		r.wg.Done()
	}()
}

func (r *runner[T, Y]) Stop() {
	select {
	case <-r.done:
	case r.sig <- struct{}{}:
	}
}

func (r *runner[T, Y]) Wait() {
	r.wg.Wait()
}

func (r *runner[T, Y]) runWorkers(w SourceWorker[T, Y]) {
	execFunc := func(ctx context.Context, id string) error {
		r.wg.Add(1)
		err := r.loop(ctx, w, id)
		r.wg.Done()
		return err
	}

	for i := 0; i < w.TargetedThreads(); i++ {
		job := base.NewJob(fmt.Sprintf("%v.%v", w.Identity(), i), execFunc)
		r.workPool.AddWorker(base.NewWorker(job))
	}

	go r.monitorWorkers(w)
}

func (r *runner[T, Y]) loop(ctx context.Context, worker SourceWorker[T, Y], id string) error {
	var err error
	uuid := utils.NewId()
	r.logger.Debug("worker spinning up",
		log.NewStringField("name", id),
		log.NewStringField("instanceId", uuid.String()),
		log.NewTField("metered", r.metrics),
	)
	err = r.workLoop(ctx, worker, id)
	r.logger.Debug("worker spinned down",
		log.NewStringField("name", id),
		log.NewStringField("instanceId", uuid.String()),
		log.NewTField("metered", r.metrics),
		log.NewErrorField(err),
	)
	return err
}

func (r *runner[T, Y]) workLoop(ctx context.Context, worker SourceWorker[T, Y], id string) error {
	if r.write {
		for {
			obj, err := r.recv(ctx)
			if err != nil {
				return err
			}
			r.recorder.RecordRecvTime(time.Since(obj.CreationTime()))
			ts := time.Now()
			res, err := worker.OnData(obj)
			r.recorder.RecordOpTime(time.Since(ts))
			ts = time.Now()
			if err == nil {
				r.itSend.Push(res)
				r.recorder.RecordOpTime(time.Since(ts))
			}
			r.recorder.RecordLoop()
		}
	}

	for {
		obj, err := r.recv(ctx)
		if err != nil {
			return err
		}
		r.recorder.RecordRecvTime(time.Since(obj.CreationTime()))
		ts := time.Now()
		worker.OnData(obj)
		r.recorder.RecordOpTime(time.Since(ts))
		r.recorder.RecordLoop()
	}
}

func (r *runner[T, Y]) recv(ctx context.Context) (T, error) {
	var t T
	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		t = r.itRecv.Next()
		if r.itRecv.Stopped() {
			return t, errors.New("stopped")
		}
	}

	return t, err
}

func (r *runner[T, Y]) monitorWorkers(w SourceWorker[T, Y]) {
	wfch := make(chan base.JobStats)
	sub := r.workPool.SubscribeJobFeed(wfch)
	defer sub.Unsubscribe()
	for {
		select {
		case feed := <-wfch:
			r.afterExitWorker(feed)
		case <-r.done:
			return
		}
	}
}
