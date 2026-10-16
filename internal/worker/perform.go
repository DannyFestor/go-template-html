package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

var (
	errShutdown     = errors.New("the worker is shutting down")
	errLeaseExpired = errors.New("the lease expired before the job finished")
	errPanicked     = errors.New("the job panicked")
)

func (w *Worker) rescue(ctx context.Context) error {
	rescued, err := w.deps.Store.RescueExpired(ctx, w.deps.Clock.Now(), errLeaseExpired.Error())
	if err != nil {
		return fmt.Errorf("rescue the jobs whose lease expired: %w", err)
	}

	if rescued > 0 {
		w.deps.Logger.WarnContext(ctx, "jobs whose lease expired were rescued", slog.Int64("rescued", rescued))
	}

	return nil
}

func (w *Worker) claim(ctx context.Context) (domain.ClaimedJob, bool, error) {
	job, found, err := w.deps.Store.Claim(ctx, w.deps.Clock.Now(), w.deps.Registry.leases())
	if err != nil {
		return domain.ClaimedJob{}, false, fmt.Errorf("claim a due job: %w", err)
	}

	return job, found, nil
}

func (w *Worker) perform(jobCtx context.Context, job domain.ClaimedJob) error {
	registered := w.deps.Registry.lookup(job.Kind)
	if job.Attempt > registered.settings.lastAttempt() {
		return w.fail(context.WithoutCancel(jobCtx), job, errLeaseExpired)
	}

	return w.record(jobCtx, job, registered.settings, handleWithinLease(jobCtx, registered, job.Args))
}

func handleWithinLease(jobCtx context.Context, registered registration, args []byte) (err error) {
	ctx, cancel := context.WithTimeout(jobCtx, registered.settings.Lease)

	defer func() {
		cancel()

		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%w: %v", errPanicked, recovered)
		}
	}()

	return registered.handle(ctx, args)
}

func (w *Worker) record(jobCtx context.Context, job domain.ClaimedJob, settings Settings, handled error) error {
	ctx := context.WithoutCancel(jobCtx)

	switch {
	case handled == nil:
		return wrapOutcome("complete", job, w.deps.Store.Complete(ctx, job))
	case errors.Is(context.Cause(jobCtx), errShutdown):
		return wrapOutcome("release", job, w.deps.Store.Release(ctx, job))
	case job.Attempt >= settings.lastAttempt():
		return w.fail(ctx, job, handled)
	default:
		return w.retry(ctx, job, settings, handled)
	}
}

func (w *Worker) retry(ctx context.Context, job domain.ClaimedJob, settings Settings, handled error) error {
	runAt := w.deps.Clock.Now().Add(settings.Backoff(job.Attempt))

	w.deps.Logger.WarnContext(ctx, "job will be retried", jobAttrs(job, handled), slog.Time("run_at", runAt))

	return wrapOutcome("retry", job, w.deps.Store.Retry(ctx, job, runAt, handled.Error()))
}

func (w *Worker) fail(ctx context.Context, job domain.ClaimedJob, handled error) error {
	w.deps.Reporter.Report(
		ctx,
		fmt.Errorf("job %s %s failed on attempt %d: %w", job.Kind, job.ID, job.Attempt, handled),
	)

	return wrapOutcome("fail", job, w.deps.Store.Fail(ctx, job, w.deps.Clock.Now(), handled.Error()))
}

func wrapOutcome(outcome string, job domain.ClaimedJob, err error) error {
	if err != nil {
		return fmt.Errorf("%s job %s %s: %w", outcome, job.Kind, job.ID, err)
	}

	return nil
}

func jobAttrs(job domain.ClaimedJob, handled error) slog.Attr {
	return slog.Group("job",
		slog.String("id", string(job.ID)),
		slog.String("kind", job.Kind),
		slog.Int("attempt", job.Attempt),
		slog.String("error", handled.Error()),
	)
}
