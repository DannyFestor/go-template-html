//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/uuidv7"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
	"github.com/DannyFestor/go-template-html/internal/worker"
)

const probeKind = "probe"

var errProbe = errors.New("probe failed")

type probeJob struct {
	Message string `json:"message"`
}

func (probeJob) Kind() string {
	return probeKind
}

type probeHandler struct {
	mutex   sync.Mutex
	handled []probeJob
	err     error
}

func (h *probeHandler) Handle(_ context.Context, job probeJob) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.handled = append(h.handled, job)

	return h.err
}

func (h *probeHandler) messages() []string {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	messages := make([]string, 0, len(h.handled))
	for _, job := range h.handled {
		messages = append(messages, job.Message)
	}

	return messages
}

type spyReporter struct {
	mutex    sync.Mutex
	reported []error
}

func (r *spyReporter) Report(_ context.Context, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.reported = append(r.reported, err)
}

func (r *spyReporter) errors() []error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return append([]error(nil), r.reported...)
}

type queueFixture struct {
	database testdb.Database
	pool     *postgres.Pool
	queue    *postgres.JobQueue
	clock    *testClock
	handler  *probeHandler
	reporter *spyReporter
	worker   *worker.Worker
}

func TestAJobEnqueuedInACommittedTransactionRunsOnce(t *testing.T) {
	t.Parallel()

	fixture := newQueueFixture(t, worker.DefaultSettings())

	err := postgres.NewTransactor(fixture.pool).WithinTransaction(t.Context(), func(ctx context.Context) error {
		return fixture.queue.Enqueue(ctx, probeJob{Message: "committed"})
	})
	require.NoError(t, err)

	fixture.runDue(t)
	fixture.runDue(t)

	assert.Equal(t, []string{"committed"}, fixture.handler.messages())
	assert.Zero(t, fixture.jobCount(t), "a completed job is deleted")
}

func TestAJobEnqueuedInARolledBackTransactionNeverRuns(t *testing.T) {
	t.Parallel()

	fixture := newQueueFixture(t, worker.DefaultSettings())

	err := postgres.NewTransactor(fixture.pool).WithinTransaction(t.Context(), func(ctx context.Context) error {
		require.NoError(t, fixture.queue.Enqueue(ctx, probeJob{Message: "rolled back"}))

		return errCallback
	})
	require.ErrorIs(t, err, errCallback)

	fixture.runDue(t)

	assert.Empty(t, fixture.handler.messages())
}

func TestAFailingJobIsRetriedWithBackoffUntilItBecomesAFailedJob(t *testing.T) {
	t.Parallel()

	settings := worker.DefaultSettings()
	settings.Retries = 2
	fixture := newQueueFixture(t, settings)
	fixture.handler.err = errProbe
	fixture.enqueue(t, "flaky")

	fixture.runDue(t)
	fixture.clock.advance(16*time.Second - time.Millisecond)
	fixture.runDue(t)
	require.Len(t, fixture.handler.messages(), 1, "the first retry waits 1⁴ + 15 seconds")

	fixture.clock.advance(20*time.Second + time.Millisecond)
	fixture.runDue(t)
	fixture.clock.advance(31*time.Second - time.Millisecond)
	fixture.runDue(t)
	require.Len(t, fixture.handler.messages(), 2, "the second retry waits 2⁴ + 15 seconds")

	fixture.clock.advance(30*time.Second + time.Millisecond)
	fixture.runDue(t)
	fixture.clock.advance(24 * time.Hour)
	fixture.runDue(t)

	assert.Len(t, fixture.handler.messages(), 3, "a failed job never runs again")
	assert.Equal(t, []string{"failed probe failed"}, fixture.jobStates(t), "a failed job is kept with its last error")
	require.Len(t, fixture.reporter.errors(), 1)
	assert.ErrorIs(t, fixture.reporter.errors()[0], errProbe)
}

func TestAJobWhoseLeaseExpiredIsRescuedAndRunAgain(t *testing.T) {
	t.Parallel()

	settings := worker.DefaultSettings()
	settings.Lease = time.Minute
	fixture := newQueueFixture(t, settings)
	fixture.enqueue(t, "abandoned")

	_, found, err := fixture.queue.Claim(t.Context(), fixture.clock.Now(), []domain.JobLease{
		{Kind: probeKind, Duration: time.Minute},
	})
	require.NoError(t, err)
	require.True(t, found)

	fixture.clock.advance(time.Minute - time.Millisecond)
	fixture.runDue(t)
	require.Empty(t, fixture.handler.messages(), "the lease still holds")

	fixture.clock.advance(time.Millisecond)
	fixture.runDue(t)

	assert.Equal(t, []string{"abandoned"}, fixture.handler.messages())
	assert.Zero(t, fixture.jobCount(t))
}

func TestAJobWhoseLeaseExpiredOnItsLastAttemptBecomesAFailedJob(t *testing.T) {
	t.Parallel()

	settings := worker.DefaultSettings()
	settings.Retries = 0
	fixture := newQueueFixture(t, settings)
	fixture.enqueue(t, "abandoned")

	_, _, err := fixture.queue.Claim(t.Context(), fixture.clock.Now(), []domain.JobLease{
		{Kind: probeKind, Duration: time.Minute},
	})
	require.NoError(t, err)

	fixture.clock.advance(time.Minute)
	fixture.runDue(t)

	assert.Empty(t, fixture.handler.messages())
	assert.Equal(t, []string{"failed the lease expired before the job finished"}, fixture.jobStates(t))
}

func TestConcurrentClaimsNeverShareAJob(t *testing.T) {
	t.Parallel()

	const jobs = 8

	fixture := newQueueFixture(t, worker.DefaultSettings())
	for range jobs {
		fixture.enqueue(t, "contended")
	}

	claimed := make(chan string, jobs)

	var claimers sync.WaitGroup
	for range jobs {
		claimers.Go(func() { claimed <- fixture.claimedIDOrError(t) })
	}

	claimers.Wait()
	close(claimed)

	ids := map[string]bool{}
	for id := range claimed {
		ids[id] = true
	}

	assert.Len(t, ids, jobs)
	assert.NotContains(t, ids, "", "a claim found no job")
}

func TestAWorkerClaimsOnlyTheKindsItHandles(t *testing.T) {
	t.Parallel()

	fixture := newQueueFixture(t, worker.DefaultSettings())
	fixture.enqueue(t, "handled")

	_, found, err := fixture.queue.Claim(t.Context(), fixture.clock.Now(), []domain.JobLease{
		{Kind: "other", Duration: time.Minute},
	})

	require.NoError(t, err)
	assert.False(t, found)
}

func newQueueFixture(t *testing.T, settings worker.Settings) *queueFixture {
	t.Helper()

	database := testdb.New(t)
	pool := open(t, database.PostgresConfig())
	clock := clockAt(time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC))
	queue := postgres.NewJobQueue(pool, uuidv7.NewGenerator(clock), clock)
	handler := &probeHandler{mutex: sync.Mutex{}, handled: nil, err: nil}
	reporter := &spyReporter{mutex: sync.Mutex{}, reported: nil}

	registry := worker.NewRegistry()
	worker.Register(registry, handler, settings)

	return &queueFixture{
		database: database,
		pool:     pool,
		queue:    queue,
		clock:    clock,
		handler:  handler,
		reporter: reporter,
		worker: worker.New(
			worker.Dependencies{
				Store:    queue,
				Registry: registry,
				Clock:    clock,
				Reporter: reporter,
				Logger:   slog.New(slog.DiscardHandler),
			},
			worker.Config{Concurrency: 1, PollInterval: time.Second, ShutdownTimeout: time.Second},
		),
	}
}

func (f *queueFixture) runDue(t *testing.T) {
	t.Helper()

	require.NoError(t, f.worker.RunDue(t.Context()))
}

func (f *queueFixture) enqueue(t *testing.T, message string) {
	t.Helper()

	require.NoError(t, f.queue.Enqueue(t.Context(), probeJob{Message: message}))
}

func (f *queueFixture) claimedIDOrError(t *testing.T) string {
	t.Helper()

	job, _, err := f.queue.Claim(
		t.Context(),
		f.clock.Now(),
		[]domain.JobLease{{Kind: probeKind, Duration: time.Minute}},
	)
	if err != nil {
		return err.Error()
	}

	return string(job.ID)
}

func (f *queueFixture) jobStates(t *testing.T) []string {
	t.Helper()

	rows, err := connect(t, f.database).Query(t.Context(), "SELECT state || ' ' || last_error FROM jobs")
	require.NoError(t, err)

	states, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	return states
}

func (f *queueFixture) jobCount(t *testing.T) int {
	t.Helper()

	var count int

	require.NoError(t, connect(t, f.database).QueryRow(t.Context(), "SELECT count(*) FROM jobs").Scan(&count))

	return count
}
