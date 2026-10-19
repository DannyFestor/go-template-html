package worker

import (
	"context"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type jobSource interface {
	Claim(ctx context.Context, now time.Time, leases []domain.JobLease) (domain.ClaimedJob, bool, error)
	RescueExpired(ctx context.Context, now time.Time, lastError string) (int64, error)
}

type jobOutcomes interface {
	Complete(ctx context.Context, job domain.ClaimedJob) error
	Retry(ctx context.Context, job domain.ClaimedJob, runAt time.Time, lastError string) error
	Fail(ctx context.Context, job domain.ClaimedJob, failedAt time.Time, lastError string) error
	Release(ctx context.Context, job domain.ClaimedJob) error
}

type periodicJobScheduler interface {
	SchedulePeriodic(ctx context.Context, job interface{ Kind() string }, scheduledFor time.Time) error
}

type jobStore interface {
	jobSource
	jobOutcomes
	periodicJobScheduler
}
