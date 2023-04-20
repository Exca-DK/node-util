package base

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

func testFunc(ctx context.Context, id string) error {
	r := rand.Intn(1000)
	duration := time.Duration(r) * time.Microsecond
	time.Sleep(duration + time.Millisecond*100)
	return nil
}

func TestWorkerPool(t *testing.T) {

	var (
		target        = 100
		targetWorkers = 10
		testWorker    = Worker{
			createdAt: time.Now(),
			job: Job{
				exec:        testFunc,
				description: "test",
			},
		}
	)

	pool := NewPool(targetWorkers)

	go pool.Start(context.TODO())
	for i := 0; i < target; i++ {
		pool.AddWorker(testWorker)
	}

	ch := make(chan JobStats, target)
	sub := pool.SubscribeJobFeed(ch)
	defer sub.Unsubscribe()
	for i := 0; i < target; i++ {
		feed := <-ch
		_ = feed
	}
}
