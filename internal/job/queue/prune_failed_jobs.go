package queue

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/action/queue"
)

type PruneFailedJobs struct{}

func (PruneFailedJobs) Kind() string {
	return "prune_failed_jobs"
}

type PruneFailedJobsHandler struct {
	prune failedJobPruner
}

func NewPruneFailedJobsHandler(prune failedJobPruner) *PruneFailedJobsHandler {
	return &PruneFailedJobsHandler{prune: prune}
}

func (h *PruneFailedJobsHandler) Handle(ctx context.Context, _ PruneFailedJobs) error {
	_, err := h.prune.Run(ctx, queue.PruneFailedJobsIn{})
	if err != nil {
		return fmt.Errorf("handle the %s job: %w", PruneFailedJobs{}.Kind(), err)
	}

	return nil
}
