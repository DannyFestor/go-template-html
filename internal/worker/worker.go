package worker

import (
	"context"
	"log/slog"
	"time"
)

type clock interface {
	Now() time.Time
}

type errorReporter interface {
	Report(ctx context.Context, err error)
}

type Dependencies struct {
	Store    jobStore
	Registry *Registry
	Clock    clock
	Reporter errorReporter
	Logger   *slog.Logger
}

type Config struct {
	Concurrency     int
	PollInterval    time.Duration
	ShutdownTimeout time.Duration
}

type Worker struct {
	deps   Dependencies
	config Config
}

func New(deps Dependencies, config Config) *Worker {
	return &Worker{deps: deps, config: config}
}

func (w *Worker) RunDue(ctx context.Context) error {
	err := w.prepareClaims(ctx)
	if err != nil {
		return err
	}

	for {
		job, found, err := w.claim(ctx)
		if err != nil || !found {
			return err
		}

		err = w.perform(ctx, job)
		if err != nil {
			return err
		}
	}
}
