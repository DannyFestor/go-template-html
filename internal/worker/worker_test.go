package worker_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/worker"
)

const (
	waitFor      = time.Second
	pollInterval = time.Millisecond
	concurrency  = 2
)

var errClaim = errors.New("claim failed")

type blockingJob struct{}

func (blockingJob) Kind() string {
	return "blocking"
}

type blockingHandler struct {
	started    chan struct{}
	release    chan struct{}
	honoursCtx bool
}

func newBlockingHandler(honoursCtx bool) *blockingHandler {
	return &blockingHandler{started: make(chan struct{}, 100), release: make(chan struct{}), honoursCtx: honoursCtx}
}

func (h *blockingHandler) Handle(ctx context.Context, _ blockingJob) error {
	h.started <- struct{}{}

	if !h.honoursCtx {
		<-h.release

		return nil
	}

	select {
	case <-h.release:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("handle the blocking job: %w", context.Cause(ctx))
	}
}

func (h *blockingHandler) awaitStart(t *testing.T) {
	t.Helper()

	select {
	case <-h.started:
	case <-time.After(waitFor):
		t.Fatal("no job started")
	}
}

type panickingJob struct{}

func (panickingJob) Kind() string {
	return "panicking"
}

type panickingHandler struct{}

func (panickingHandler) Handle(context.Context, panickingJob) error {
	panic("boom")
}

type memoryStore struct {
	mutex     sync.Mutex
	available []domain.ClaimedJob
	outcomes  map[domain.JobID]string
	claimErrs []error
}

func newMemoryStore(kind string, jobs int) *memoryStore {
	store := &memoryStore{mutex: sync.Mutex{}, available: nil, outcomes: map[domain.JobID]string{}, claimErrs: nil}
	for index := range jobs {
		store.add(domain.ClaimedJob{ID: domain.JobID(strconv.Itoa(index)), Kind: kind, Args: []byte("{}"), Attempt: 0})
	}

	return store
}

func (s *memoryStore) Claim(context.Context, time.Time, []domain.JobLease) (domain.ClaimedJob, bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	noJob := domain.ClaimedJob{ID: "", Kind: "", Args: nil, Attempt: 0}

	if len(s.claimErrs) > 0 {
		err := s.claimErrs[0]
		s.claimErrs = s.claimErrs[1:]

		return noJob, false, err
	}

	if len(s.available) == 0 {
		return noJob, false, nil
	}

	job := s.available[0]
	s.available = s.available[1:]
	job.Attempt++

	return job, true, nil
}

func (s *memoryStore) RescueExpired(context.Context, time.Time, string) (int64, error) {
	return 0, nil
}

func (s *memoryStore) Complete(_ context.Context, job domain.ClaimedJob) error {
	return s.record(job, "completed")
}

func (s *memoryStore) Retry(_ context.Context, job domain.ClaimedJob, _ time.Time, lastError string) error {
	return s.record(job, "retried: "+lastError)
}

func (s *memoryStore) Fail(_ context.Context, job domain.ClaimedJob, _ time.Time, lastError string) error {
	return s.record(job, "failed: "+lastError)
}

func (s *memoryStore) Release(_ context.Context, job domain.ClaimedJob) error {
	return s.record(job, "released")
}

func (s *memoryStore) add(job domain.ClaimedJob) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.available = append(s.available, job)
}

func (s *memoryStore) record(job domain.ClaimedJob, outcome string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.outcomes[job.ID] = outcome

	return nil
}

func (s *memoryStore) outcome(id domain.JobID) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.outcomes[id]
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
}

type spyReporter struct {
	mutex    sync.Mutex
	reported []error
}

func newSpyReporter() *spyReporter {
	return &spyReporter{mutex: sync.Mutex{}, reported: nil}
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

type running struct {
	cancel context.CancelFunc
	result chan error
}

func (r running) stopped(t *testing.T) error {
	t.Helper()

	select {
	case err := <-r.result:
		return err
	case <-time.After(waitFor):
		t.Fatal("the worker did not stop")

		return nil
	}
}

func TestRunPerformsAtMostConcurrencyJobsAtOnce(t *testing.T) {
	t.Parallel()

	handler := newBlockingHandler(false)
	store := newMemoryStore(blockingJob{}.Kind(), concurrency+1)
	run := start(t, newWorker(store, handler, newSpyReporter(), time.Minute))

	handler.awaitStart(t)
	handler.awaitStart(t)
	assert.Never(t, func() bool { return len(handler.started) > 0 }, 50*time.Millisecond, pollInterval)

	close(handler.release)
	handler.awaitStart(t)
	run.cancel()

	require.NoError(t, run.stopped(t))
	assert.Equal(t, []string{"completed", "completed", "completed"},
		[]string{store.outcome("0"), store.outcome("1"), store.outcome("2")})
}

func TestRunLetsInFlightJobsFinishOnShutdown(t *testing.T) {
	t.Parallel()

	handler := newBlockingHandler(true)
	store := newMemoryStore(blockingJob{}.Kind(), 1)
	run := start(t, newWorker(store, handler, newSpyReporter(), time.Minute))

	handler.awaitStart(t)
	run.cancel()
	store.add(domain.ClaimedJob{ID: "late", Kind: blockingJob{}.Kind(), Args: []byte("{}"), Attempt: 0})
	close(handler.release)

	require.NoError(t, run.stopped(t))
	assert.Equal(t, "completed", store.outcome("0"))
	assert.Empty(t, store.outcome("late"), "no job is claimed after shutdown began")
}

func TestRunReleasesJobsStillRunningAtTheShutdownTimeout(t *testing.T) {
	t.Parallel()

	handler := newBlockingHandler(true)
	store := newMemoryStore(blockingJob{}.Kind(), 1)
	run := start(t, newWorker(store, handler, newSpyReporter(), 10*time.Millisecond))

	handler.awaitStart(t)
	run.cancel()

	require.Error(t, run.stopped(t))
	assert.Equal(t, "released", store.outcome("0"))
}

func TestRunReportsAFailedClaimAndKeepsPolling(t *testing.T) {
	t.Parallel()

	handler := newBlockingHandler(false)
	close(handler.release)

	store := newMemoryStore(blockingJob{}.Kind(), 1)
	store.claimErrs = []error{errClaim}
	reporter := newSpyReporter()
	run := start(t, newWorker(store, handler, reporter, time.Minute))

	handler.awaitStart(t)
	run.cancel()

	require.NoError(t, run.stopped(t))
	require.Len(t, reporter.errors(), 1)
	assert.ErrorIs(t, reporter.errors()[0], errClaim)
}

func TestAPanickingJobIsRetried(t *testing.T) {
	t.Parallel()

	store := newMemoryStore(panickingJob{}.Kind(), 1)
	registry := worker.NewRegistry()
	worker.Register(registry, panickingHandler{}, worker.DefaultSettings())

	err := worker.New(dependencies(store, registry, newSpyReporter()), config(time.Minute)).RunDue(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "retried: the job panicked: boom", store.outcome("0"))
}

func TestAJobOutlastingItsLeaseIsCancelledAndRetried(t *testing.T) {
	t.Parallel()

	handler := newBlockingHandler(true)
	store := newMemoryStore(blockingJob{}.Kind(), 1)
	settings := worker.DefaultSettings()
	settings.Lease = 10 * time.Millisecond
	registry := worker.NewRegistry()
	worker.Register(registry, handler, settings)

	err := worker.New(dependencies(store, registry, newSpyReporter()), config(time.Minute)).RunDue(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "retried: handle the blocking job: context deadline exceeded", store.outcome("0"))
}

func newWorker(
	store *memoryStore,
	handler *blockingHandler,
	reporter *spyReporter,
	shutdownTimeout time.Duration,
) *worker.Worker {
	registry := worker.NewRegistry()
	worker.Register(registry, handler, worker.DefaultSettings())

	return worker.New(dependencies(store, registry, reporter), config(shutdownTimeout))
}

func dependencies(store *memoryStore, registry *worker.Registry, reporter *spyReporter) worker.Dependencies {
	return worker.Dependencies{
		Store:    store,
		Registry: registry,
		Clock:    fixedClock{},
		Reporter: reporter,
		Logger:   slog.New(slog.DiscardHandler),
	}
}

func config(shutdownTimeout time.Duration) worker.Config {
	return worker.Config{Concurrency: concurrency, PollInterval: pollInterval, ShutdownTimeout: shutdownTimeout}
}

func start(t *testing.T, w *worker.Worker) running {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	run := running{cancel: cancel, result: make(chan error, 1)}

	go func() { run.result <- w.Run(ctx) }()

	return run
}
