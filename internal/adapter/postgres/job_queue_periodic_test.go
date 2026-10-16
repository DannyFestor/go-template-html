//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/worker"
)

func TestTwoWorkersScheduleOnePeriodicRunPerInterval(t *testing.T) {
	t.Parallel()

	fixture := newQueueFixture(t, worker.DefaultSettings())
	workers := []*worker.Worker{fixture.newPeriodicWorker(time.Hour), fixture.newPeriodicWorker(time.Hour)}

	runDueOnEach(t, workers)
	fixture.clock.advance(time.Hour - time.Millisecond)
	runDueOnEach(t, workers)
	require.Empty(t, fixture.handler.messages(), "the first run waits for the next full hour")

	fixture.clock.advance(time.Millisecond)
	runDueOnEach(t, workers)
	require.Len(t, fixture.handler.messages(), 1)

	fixture.clock.advance(time.Hour)
	runDueOnEach(t, workers)
	assert.Len(t, fixture.handler.messages(), 2)
}

func (f *queueFixture) newPeriodicWorker(interval time.Duration) *worker.Worker {
	registry := worker.NewRegistry()
	worker.RegisterPeriodic(registry, f.handler, worker.DefaultSettings(), interval)

	return f.newWorker(registry)
}

func runDueOnEach(t *testing.T, workers []*worker.Worker) {
	t.Helper()

	for _, each := range workers {
		require.NoError(t, each.RunDue(t.Context()))
	}
}
