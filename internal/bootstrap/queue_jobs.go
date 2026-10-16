package bootstrap

import (
	"time"

	"github.com/DannyFestor/go-template-html/internal/action/queue"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	queuejob "github.com/DannyFestor/go-template-html/internal/job/queue"
	"github.com/DannyFestor/go-template-html/internal/worker"
)

const pruneFailedJobsInterval = 24 * time.Hour

func registerQueueJobs(jobs *worker.Registry, jobQueue *postgres.JobQueue, config Queue) {
	prune := queue.NewPruneFailedJobs(jobQueue, clock.System{}, config.FailedRetention)

	worker.RegisterPeriodic(
		jobs,
		queuejob.NewPruneFailedJobsHandler(prune),
		worker.DefaultSettings(),
		pruneFailedJobsInterval,
	)
}
