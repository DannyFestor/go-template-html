//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/worker"
)

func TestFailedJobsThatFailedBeforeTheCutoffArePrunedAndNewerOnesStay(t *testing.T) {
	t.Parallel()

	settings := worker.DefaultSettings()
	settings.Retries = 0
	fixture := newQueueFixture(t, settings)
	fixture.handler.err = errProbe

	fixture.enqueue(t, "failed before the cutoff")
	fixture.runDue(t)
	cutoff := fixture.clock.Now().Add(time.Millisecond)
	fixture.clock.advance(time.Millisecond)
	fixture.enqueue(t, "failed at the cutoff")
	fixture.runDue(t)
	fixture.enqueue(t, "still queued")

	require.NoError(t, fixture.queue.DeleteFailedBefore(t.Context(), cutoff))

	assert.Equal(t, []string{"failed at the cutoff", "still queued"}, fixture.remainingMessages(t))
}

func (f *queueFixture) remainingMessages(t *testing.T) []string {
	t.Helper()

	rows, err := connect(t, f.database).Query(t.Context(), "SELECT args->>'message' FROM jobs ORDER BY 1")
	require.NoError(t, err)

	messages, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	return messages
}
