package runner

import "time"

type WorkItem interface {
	CreationTime() time.Time
}

type Runner[T any, Y any] interface {
	Start(SourceWorker[T, Y])
	Stop()
	Wait()
}

type IteratorRecv[T any] interface {
	Next() T
	Stopped() bool
}

type IteratorSend[T any] interface {
	Push(T)
	Stop()
}

type SourceWorker[T any, Y any] interface {
	OnData(T) (Y, error)
	TargetedThreads() int
	Identity() string
}
