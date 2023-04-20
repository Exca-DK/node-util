package base

import (
	"context"
	"time"
)

const (
	empty string = "empty"
)

func NewJob(description string, method func(ctx context.Context, id string) error) Job {
	return Job{exec: method, description: description}
}

type Job struct {
	exec        func(ctx context.Context, id string) error
	description string
}

type JobStats struct {
	JobDescription string
	JobError       error
	CreatedAt      time.Time
	StartedAt      time.Time
	FinishedAt     time.Time
}

func NewWorker(job Job) Worker {
	return Worker{createdAt: time.Now(), job: job}
}

func emptyWorker() Worker {
	return Worker{createdAt: time.Now(), job: Job{
		exec: func(ctx context.Context, id string) error {
			return nil
		},
		description: empty,
	}}
}

type Worker struct {
	createdAt time.Time
	job       Job
}

func (w Worker) Run(ctx context.Context) JobStats {
	startedAt := time.Now()
	err := w.job.exec(ctx, w.job.description)

	return JobStats{
		JobDescription: w.job.description,
		JobError:       err,
		CreatedAt:      w.createdAt,
		StartedAt:      startedAt,
		FinishedAt:     time.Now(),
	}
}
