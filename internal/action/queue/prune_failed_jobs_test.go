package queue_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/queue"
)

func TestPruneFailedJobsDeletesTheJobsThatFailedLongerThanTheRetentionAgo(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	clock := queue.NewMockClock(controller)
	clock.EXPECT().Now().Return(time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC))

	jobs := queue.NewMockFailedJobDeleter(controller)
	jobs.EXPECT().DeleteFailedBefore(gomock.Any(), time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC))

	_, err := queue.NewPruneFailedJobs(jobs, clock, 7*24*time.Hour).Run(t.Context(), queue.PruneFailedJobsIn{})

	require.NoError(t, err)
}
