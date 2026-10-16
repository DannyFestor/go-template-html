package queue

import (
	"context"
	"fmt"
	"time"
)

type PruneFailedJobs struct {
	jobs      failedJobDeleter
	clock     clock
	retention time.Duration
}

type PruneFailedJobsIn struct{}

type PruneFailedJobsOut struct{}

func NewPruneFailedJobs(jobs failedJobDeleter, clock clock, retention time.Duration) *PruneFailedJobs {
	return &PruneFailedJobs{jobs: jobs, clock: clock, retention: retention}
}

func (p *PruneFailedJobs) Run(ctx context.Context, _ PruneFailedJobsIn) (PruneFailedJobsOut, error) {
	err := p.jobs.DeleteFailedBefore(ctx, p.clock.Now().Add(-p.retention))
	if err != nil {
		return PruneFailedJobsOut{}, fmt.Errorf("prune the failed jobs older than %s: %w", p.retention, err)
	}

	return PruneFailedJobsOut{}, nil
}
