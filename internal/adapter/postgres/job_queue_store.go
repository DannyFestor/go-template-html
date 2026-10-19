package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

func (q *JobQueue) Claim(
	ctx context.Context,
	now time.Time,
	leases []domain.JobLease,
) (domain.ClaimedJob, bool, error) {
	row, err := q.pool.queries(ctx).ClaimJob(ctx, claimParams(now, leases))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ClaimedJob{ID: "", Kind: "", Args: nil, Attempt: 0}, false, nil
	}

	if err != nil {
		return domain.ClaimedJob{}, false, fmt.Errorf("lease the next available job: %w", err)
	}

	return domain.ClaimedJob{
		ID:      domain.JobID(row.ID.String()),
		Kind:    row.Kind,
		Args:    row.Args,
		Attempt: int(row.Attempt),
	}, true, nil
}

func claimParams(now time.Time, leases []domain.JobLease) sqlcgen.ClaimJobParams {
	params := sqlcgen.ClaimJobParams{
		Now:     now,
		Kinds:   make([]string, 0, len(leases)),
		LeaseMs: make([]int64, 0, len(leases)),
	}

	for _, lease := range leases {
		params.Kinds = append(params.Kinds, lease.Kind)
		params.LeaseMs = append(params.LeaseMs, lease.Duration.Milliseconds())
	}

	return params
}

func (q *JobQueue) RescueExpired(ctx context.Context, now time.Time, lastError string) (int64, error) {
	rescued, err := q.pool.queries(ctx).RescueExpiredJobs(ctx, sqlcgen.RescueExpiredJobsParams{
		LastError: lastError,
		Now:       now,
	})
	if err != nil {
		return 0, fmt.Errorf("make the expired leases available: %w", err)
	}

	return rescued, nil
}

func (q *JobQueue) DeleteFailedBefore(ctx context.Context, failedBefore time.Time) error {
	err := q.pool.queries(ctx).DeleteFailedJobsBefore(ctx, failedBefore)
	if err != nil {
		return fmt.Errorf("delete the jobs that failed before %s: %w", failedBefore, err)
	}

	return nil
}

func (q *JobQueue) Complete(ctx context.Context, job domain.ClaimedJob) error {
	return q.update(ctx, job, func(queries *sqlcgen.Queries, ref jobRef) error {
		return queries.CompleteJob(ctx, sqlcgen.CompleteJobParams{ID: ref.id, Attempt: ref.attempt})
	})
}

func (q *JobQueue) Retry(ctx context.Context, job domain.ClaimedJob, runAt time.Time, lastError string) error {
	return q.update(ctx, job, func(queries *sqlcgen.Queries, ref jobRef) error {
		return queries.RetryJob(ctx, sqlcgen.RetryJobParams{
			RunAt:     runAt,
			LastError: lastError,
			ID:        ref.id,
			Attempt:   ref.attempt,
		})
	})
}

func (q *JobQueue) Fail(ctx context.Context, job domain.ClaimedJob, failedAt time.Time, lastError string) error {
	return q.update(ctx, job, func(queries *sqlcgen.Queries, ref jobRef) error {
		return queries.FailJob(ctx, sqlcgen.FailJobParams{
			FailedAt:  failedAt,
			LastError: lastError,
			ID:        ref.id,
			Attempt:   ref.attempt,
		})
	})
}

func (q *JobQueue) Release(ctx context.Context, job domain.ClaimedJob) error {
	return q.update(ctx, job, func(queries *sqlcgen.Queries, ref jobRef) error {
		return queries.ReleaseJob(ctx, sqlcgen.ReleaseJobParams{ID: ref.id, Attempt: ref.attempt})
	})
}

func (q *JobQueue) update(
	ctx context.Context,
	job domain.ClaimedJob,
	query func(queries *sqlcgen.Queries, ref jobRef) error,
) error {
	ref, err := newJobRef(job)
	if err != nil {
		return err
	}

	err = query(q.pool.queries(ctx), ref)
	if err != nil {
		return fmt.Errorf("update the %s job %s: %w", job.Kind, job.ID, err)
	}

	return nil
}
