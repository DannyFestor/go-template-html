package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

func NewJobQueue(pool *Pool, ids idGenerator, clock clock) *JobQueue {
	return &JobQueue{pool: pool, ids: ids, clock: clock}
}

func (q *JobQueue) Enqueue(ctx context.Context, job interface{ Kind() string }) error {
	args, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("encode the %s job: %w", job.Kind(), err)
	}

	id, err := parseJobID(domain.JobID(q.ids.NewID()))
	if err != nil {
		return err
	}

	err = q.pool.queries(ctx).InsertJob(ctx, sqlcgen.InsertJobParams{
		ID:    id,
		Kind:  job.Kind(),
		Args:  args,
		RunAt: q.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("enqueue the %s job: %w", job.Kind(), err)
	}

	return nil
}
