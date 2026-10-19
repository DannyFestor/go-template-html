package worker

import (
	"context"
	"fmt"
	"time"
)

type periodicJob struct {
	job      job
	interval time.Duration
}

// The next boundary rather than the current one: a completed job is deleted, so a worker scheduling the current
// boundary after another worker ran it would run it a second time.
func (p periodicJob) nextRun(now time.Time) time.Time {
	return now.UTC().Truncate(p.interval).Add(p.interval)
}

func (w *Worker) schedule(ctx context.Context) error {
	now := w.deps.Clock.Now()

	for _, periodic := range w.deps.Registry.periodicJobs() {
		err := w.deps.Store.SchedulePeriodic(ctx, periodic.job, periodic.nextRun(now))
		if err != nil {
			return fmt.Errorf("schedule the periodic %s job: %w", periodic.job.Kind(), err)
		}
	}

	return nil
}
