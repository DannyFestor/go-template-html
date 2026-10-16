package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"uuid"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type idGenerator interface {
	NewID() string
}

type clock interface {
	Now() time.Time
}

type JobQueue struct {
	pool  *Pool
	ids   idGenerator
	clock clock
}

type encodedJob struct {
	id   uuid.UUID
	args []byte
}

func NewJobQueue(pool *Pool, ids idGenerator, clock clock) *JobQueue {
	return &JobQueue{pool: pool, ids: ids, clock: clock}
}

func (q *JobQueue) Enqueue(ctx context.Context, job interface{ Kind() string }) error {
	encoded, err := q.encode(job)
	if err != nil {
		return err
	}

	err = q.pool.queries(ctx).InsertJob(ctx, sqlcgen.InsertJobParams{
		ID:    encoded.id,
		Kind:  job.Kind(),
		Args:  encoded.args,
		RunAt: q.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("enqueue the %s job: %w", job.Kind(), err)
	}

	return nil
}

func (q *JobQueue) SchedulePeriodic(ctx context.Context, job interface{ Kind() string }, scheduledFor time.Time) error {
	encoded, err := q.encode(job)
	if err != nil {
		return err
	}

	err = q.pool.queries(ctx).InsertPeriodicJob(ctx, sqlcgen.InsertPeriodicJobParams{
		ID:           encoded.id,
		Kind:         job.Kind(),
		Args:         encoded.args,
		ScheduledFor: scheduledFor,
		CreatedAt:    q.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("schedule the %s job for %s: %w", job.Kind(), scheduledFor, err)
	}

	return nil
}

func (q *JobQueue) encode(job interface{ Kind() string }) (encodedJob, error) {
	args, err := json.Marshal(job)
	if err != nil {
		return encodedJob{}, fmt.Errorf("encode the %s job: %w", job.Kind(), err)
	}

	id, err := parseJobID(domain.JobID(q.ids.NewID()))
	if err != nil {
		return encodedJob{}, err
	}

	return encodedJob{id: id, args: args}, nil
}
