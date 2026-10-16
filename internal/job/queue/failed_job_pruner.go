package queue

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/action/queue"
)

//go:generate go tool mockgen -source=failed_job_pruner.go -destination=mock_failed_job_pruner_test.go -package=queue -mock_names=failedJobPruner=MockFailedJobPruner

type failedJobPruner interface {
	Run(ctx context.Context, in queue.PruneFailedJobsIn) (queue.PruneFailedJobsOut, error)
}
