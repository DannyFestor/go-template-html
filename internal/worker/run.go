package worker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var errDrainTimedOut = errors.New("jobs were still running, so they were cancelled and released")

func (w *Worker) Run(ctx context.Context) error {
	jobCtx, stopJobs := context.WithCancelCause(context.WithoutCancel(ctx))
	defer stopJobs(nil)

	running := newPool(w.config.Concurrency)
	w.poll(ctx, jobCtx, running)

	return w.drain(running, stopJobs)
}

func (w *Worker) poll(ctx, jobCtx context.Context, running *pool) {
	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for ctx.Err() == nil {
		w.fill(ctx, jobCtx, running)

		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
}

func (w *Worker) fill(ctx, jobCtx context.Context, running *pool) {
	w.report(ctx, w.prepareClaims(ctx))

	for ctx.Err() == nil && running.tryAcquire() {
		job, found, err := w.claim(jobCtx)
		if err != nil || !found {
			running.release()
			w.report(ctx, err)

			return
		}

		running.start(func() { w.report(jobCtx, w.perform(jobCtx, job)) })
	}
}

func (w *Worker) drain(running *pool, stopJobs context.CancelCauseFunc) error {
	if running.waitFor(w.config.ShutdownTimeout) {
		return nil
	}

	stopJobs(errShutdown)
	running.wait()

	return fmt.Errorf("drain within %s: %w", w.config.ShutdownTimeout, errDrainTimedOut)
}

func (w *Worker) report(ctx context.Context, err error) {
	if err == nil || errors.Is(err, context.Canceled) && ctx.Err() != nil {
		return
	}

	w.deps.Reporter.Report(ctx, err)
}
