package health_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/handler/health"
)

const ttl = 5 * time.Second

var errDown = errors.New("connection refused")

type fakePinger struct {
	pings atomic.Int32
	down  atomic.Bool
	delay time.Duration
}

func newFakePinger(delay time.Duration) *fakePinger {
	return &fakePinger{pings: atomic.Int32{}, down: atomic.Bool{}, delay: delay}
}

func (f *fakePinger) Ping(ctx context.Context) error {
	f.pings.Add(1)

	select {
	case <-time.After(f.delay):
	case <-ctx.Done():
		return fmt.Errorf("ping: %w", ctx.Err())
	}

	if f.down.Load() {
		return errDown
	}

	return nil
}

func TestReadinessReportsOKWhileThePingSucceeds(t *testing.T) {
	t.Parallel()

	recorder := serveReadiness(t, newFakePinger(0))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "ok\n", recorder.Body.String())
}

func TestReadinessReportsUnavailableWhileThePingFails(t *testing.T) {
	t.Parallel()

	pinger := newFakePinger(0)
	pinger.down.Store(true)

	recorder := serveReadiness(t, pinger)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Equal(t, "unavailable\n", recorder.Body.String())
}

func TestDependenciesFailWhileAnyOneIsDown(t *testing.T) {
	t.Parallel()

	up, down := newFakePinger(0), newFakePinger(0)
	down.down.Store(true)

	require.NoError(t, health.NewDependencies(up, up).Ping(t.Context()))
	require.ErrorIs(t, health.NewDependencies(up, down).Ping(t.Context()), errDown)
	assert.ErrorIs(t, health.NewDependencies(down, up).Ping(t.Context()), errDown)
}

func TestCachedPingPingsAtMostOncePerTTL(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		pinger := newFakePinger(0)
		cached := health.NewCachedPing(pinger, ttl)

		require.NoError(t, cached.Ping(t.Context()))
		time.Sleep(ttl - time.Nanosecond)
		require.NoError(t, cached.Ping(t.Context()))

		assert.Equal(t, int32(1), pinger.pings.Load())
	})
}

func TestCachedPingReportsRecoveryOnceTheTTLRunsOut(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		pinger := newFakePinger(0)
		pinger.down.Store(true)
		cached := health.NewCachedPing(pinger, ttl)

		require.ErrorIs(t, cached.Ping(t.Context()), errDown)

		pinger.down.Store(false)
		require.ErrorIs(t, cached.Ping(t.Context()), errDown)

		time.Sleep(ttl)
		assert.NoError(t, cached.Ping(t.Context()))
	})
}

func TestCachedPingSharesOnePingBetweenConcurrentCallers(t *testing.T) {
	t.Parallel()

	pinger := newFakePinger(10 * time.Millisecond)
	cached := health.NewCachedPing(pinger, time.Minute)

	var callers sync.WaitGroup
	for range 10 {
		callers.Go(func() { assert.NoError(t, cached.Ping(t.Context())) })
	}

	callers.Wait()

	assert.Equal(t, int32(1), pinger.pings.Load())
}

func TestCachedPingFailsAPingThatHangs(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		cached := health.NewCachedPing(newFakePinger(time.Hour), ttl)

		assert.ErrorIs(t, cached.Ping(t.Context()), context.DeadlineExceeded)
	})
}

func TestCachedPingOutlivesTheRequestThatStartedIt(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		cached := health.NewCachedPing(newFakePinger(0), ttl)

		canceled, cancel := context.WithCancel(t.Context())
		cancel()

		assert.NoError(t, cached.Ping(canceled))
	})
}

func serveReadiness(t *testing.T, pinger *fakePinger) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()

	health.NewReadiness(pinger).ServeHTTP(
		recorder,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil),
	)

	return recorder
}
