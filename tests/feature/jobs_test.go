//go:build feature

package feature_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAJobARequestEnqueuedRunsWhenTheTestRunsTheQueuedJobs(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	enqueued := running.send(t, outgoing{method: http.MethodPost, path: probeEnqueue, headers: nil, body: nil})
	require.Equal(t, http.StatusNoContent, enqueued.status)
	require.Empty(t, running.probeJobs.handled(), "a queued job waits for the worker")

	running.RunQueuedJobs(t)
	running.RunQueuedJobs(t)

	assert.Equal(t, []string{probeJobMessage}, running.probeJobs.handled())
}
