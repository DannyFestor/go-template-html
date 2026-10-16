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

func hitLoginLimit(t *testing.T, limiter *fakeRateLimiter) error {
	t.Helper()

	limit := web.NewRateLimit(limiter, web.Limit{Name: "login", MaxAttempts: 5, Window: time.Minute})

	return limit.Hit(t.Context(), "ada@example.com|192.0.2.1") //nolint:wrapcheck // the tests inspect the error itself
}

func TestRateLimitCountsTheAttemptUnderTheLimitsName(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: true, retryAfter: 0, err: nil, hits: nil}

	require.NoError(t, hitLoginLimit(t, limiter))
	assert.Equal(t, []hit{{key: "login:ada@example.com|192.0.2.1", maxAttempts: 5, window: time.Minute}}, limiter.hits)
}

func TestRateLimitRefusesARefusedAttemptWithTooManyRequests(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: false, retryAfter: 42 * time.Second, err: nil, hits: nil}

	assert.Equal(t, web.ErrTooManyRequests(42*time.Second), hitLoginLimit(t, limiter))
}

func TestRateLimitRefusesARefusedAttemptEvenWithNoWaitLeft(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: false, retryAfter: 0, err: nil, hits: nil}

	assert.Equal(t, web.ErrTooManyRequests(0), hitLoginLimit(t, limiter))
}

func TestRateLimitRefusesWithAReportedServiceUnavailableWhenItCannotCount(t *testing.T) {
	t.Parallel()

	limiter := &fakeRateLimiter{allowed: true, retryAfter: 0, err: errValkeyDown, hits: nil}

	response := respondWithError(t, nil, false, hitLoginLimit(t, limiter))

	assert.Equal(t, http.StatusServiceUnavailable, response.recorder.Code)
	assert.Equal(t, "Service Unavailable", response.main(t).Find("h1").Text())
	require.Len(t, response.reporter.reported, 1)
	assert.ErrorIs(t, response.reporter.reported[0], errValkeyDown)
}
