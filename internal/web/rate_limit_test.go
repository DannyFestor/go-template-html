package web_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/web"
)

var errValkeyDown = errors.New("valkey down")

type hit struct {
	key         string
	maxAttempts int
	window      time.Duration
}

type fakeRateLimiter struct {
	allowed    bool
	retryAfter time.Duration
	err        error
	hits       []hit
	cleared    []string
}

func (f *fakeRateLimiter) Hit(
	_ context.Context,
	key string,
	maxAttempts int,
	window time.Duration,
) (bool, time.Duration, error) {
	f.hits = append(f.hits, hit{key: key, maxAttempts: maxAttempts, window: window})

	return f.allowed, f.retryAfter, f.err
}

func (f *fakeRateLimiter) Clear(_ context.Context, key string) error {
	f.cleared = append(f.cleared, key)

	return f.err
}

func loginLimit(limiter *fakeRateLimiter) *web.RateLimit {
	return web.NewRateLimit(limiter, web.Limit{Name: "login", MaxAttempts: 5, Window: time.Minute})
}

func hitLoginLimit(t *testing.T, limiter *fakeRateLimiter) error {
	t.Helper()

	limit := loginLimit(limiter)

	return limit.Hit(t.Context(), "ada@example.com|192.0.2.1") //nolint:wrapcheck // the tests inspect the error itself
}

func TestRateLimitClearsTheKeyUnderTheLimitsName(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: true, retryAfter: 0, err: nil, hits: nil, cleared: nil}

	require.NoError(t, loginLimit(limiter).Clear(t.Context(), "ada@example.com|192.0.2.1"))
	assert.Equal(t, []string{"login:ada@example.com|192.0.2.1"}, limiter.cleared)
}

func TestRateLimitFailsWithAServiceUnavailableWhenItCannotClear(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: true, retryAfter: 0, err: errValkeyDown, hits: nil, cleared: nil}

	err := loginLimit(limiter).Clear(t.Context(), "ada@example.com|192.0.2.1")

	statusErr, ok := errors.AsType[*web.StatusError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusServiceUnavailable, statusErr.Status)
	assert.ErrorIs(t, err, errValkeyDown)
}

func TestRateLimitCountsTheAttemptUnderTheLimitsName(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: true, retryAfter: 0, err: nil, hits: nil, cleared: nil}

	require.NoError(t, hitLoginLimit(t, limiter))
	assert.Equal(t, []hit{{key: "login:ada@example.com|192.0.2.1", maxAttempts: 5, window: time.Minute}}, limiter.hits)
}

func TestRateLimitRefusesARefusedAttemptWithTooManyRequests(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: false, retryAfter: 42 * time.Second, err: nil, hits: nil, cleared: nil}

	assert.Equal(t, web.ErrTooManyRequests(42*time.Second), hitLoginLimit(t, limiter))
}

func TestRateLimitRefusesARefusedAttemptEvenWithNoWaitLeft(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: false, retryAfter: 0, err: nil, hits: nil, cleared: nil}

	assert.Equal(t, web.ErrTooManyRequests(0), hitLoginLimit(t, limiter))
}

func TestRateLimitRefusesWithAReportedServiceUnavailableWhenItCannotCount(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: true, retryAfter: 0, err: errValkeyDown, hits: nil, cleared: nil}

	response := respondWithError(t, nil, false, hitLoginLimit(t, limiter))

	assert.Equal(t, http.StatusServiceUnavailable, response.recorder.Code)
	assert.Equal(t, "Service Unavailable", response.main(t).Find("h1").Text())
	require.Len(t, response.reporter.reported, 1)
	assert.ErrorIs(t, response.reporter.reported[0], errValkeyDown)
}
