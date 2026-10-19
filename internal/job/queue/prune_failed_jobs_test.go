package queue_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/queue"
	queuejob "github.com/DannyFestor/go-template-html/internal/job/queue"
)

var errPrune = errors.New("prune failed")

func TestAFailedPruneFailsTheJobSoItIsRetried(t *testing.T) {
	t.Parallel()

	prune := queuejob.NewMockFailedJobPruner(gomock.NewController(t))
	prune.EXPECT().Run(gomock.Any(), queue.PruneFailedJobsIn{}).Return(queue.PruneFailedJobsOut{}, errPrune)

	err := queuejob.NewPruneFailedJobsHandler(prune).Handle(t.Context(), queuejob.PruneFailedJobs{})

	require.ErrorIs(t, err, errPrune)
}
