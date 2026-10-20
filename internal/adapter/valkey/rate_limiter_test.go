//go:build integration

package valkey_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

const (
	rateLimitKey   = "login:ada@example.com|192.0.2.1"
	maxAttempts    = 3
	shortWindow    = 400 * time.Millisecond
	windowQuarters = 4
	expiryMargin   = 100 * time.Millisecond
)

func TestAttemptsWithinTheLimitAreAllowed(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(t, testvalkey.New(t))

	for range maxAttempts {
		assertAllowed(t, limiter, rateLimitKey, time.Minute)
	}
}

func TestTheAttemptOverTheLimitIsRefusedUntilTheWindowEnds(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(t, testvalkey.New(t))
	hitTimes(t, limiter, maxAttempts, time.Minute)

	retryAfter := assertRefused(t, limiter, time.Minute)

	assert.Greater(t, retryAfter, 59*time.Second)
	assert.LessOrEqual(t, retryAfter, time.Minute)
}

func TestAttemptsAreAllowedAgainOnceTheWindowEnds(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(t, testvalkey.New(t))
	hitTimes(t, limiter, maxAttempts+1, shortWindow)

	time.Sleep(shortWindow + expiryMargin)

	assertAllowed(t, limiter, rateLimitKey, shortWindow)
}

func TestRefusedAttemptsDoNotExtendTheWindow(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(t, testvalkey.New(t))
	hitTimes(t, limiter, maxAttempts, shortWindow)

	for range windowQuarters - 1 {
		time.Sleep(shortWindow / windowQuarters)
		assertRefused(t, limiter, shortWindow)
	}

	time.Sleep(shortWindow/windowQuarters + expiryMargin)

	assertAllowed(t, limiter, rateLimitKey, shortWindow)
}

func TestAFreshWindowAllowsItsFullLimitRightAfterAnEdge(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(t, testvalkey.New(t))
	hitTimes(t, limiter, maxAttempts, shortWindow)

	time.Sleep(shortWindow + expiryMargin)

	for range maxAttempts {
		assertAllowed(t, limiter, rateLimitKey, shortWindow)
	}

	assertRefused(t, limiter, shortWindow)
}

func TestEachKeyIsCountedOnItsOwn(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(t, testvalkey.New(t))
	hitTimes(t, limiter, maxAttempts+1, time.Minute)

	assertAllowed(t, limiter, "login:grace@example.com|192.0.2.1", time.Minute)
}

func TestAttemptsAreIsolatedByKeyPrefix(t *testing.T) {
	t.Parallel()

	hitTimes(t, newRateLimiter(t, testvalkey.New(t)), maxAttempts+1, time.Minute)

	assertAllowed(t, newRateLimiter(t, testvalkey.New(t)), rateLimitKey, time.Minute)
}

func TestClearingAKeyStartsItsCountAfresh(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(t, testvalkey.New(t))
	hitTimes(t, limiter, maxAttempts+1, time.Minute)

	require.NoError(t, limiter.Clear(t.Context(), rateLimitKey))

	for range maxAttempts {
		assertAllowed(t, limiter, rateLimitKey, time.Minute)
	}
}

func TestClearingFailsWhileValkeyIsUnreachable(t *testing.T) {
	t.Parallel()

	assert.Error(t, newRateLimiter(t, testvalkey.Unreachable(t)).Clear(t.Context(), rateLimitKey))
}

func TestHittingFailsWhileValkeyIsUnreachable(t *testing.T) {
	t.Parallel()

	_, _, err := newRateLimiter(t, testvalkey.Unreachable(t)).Hit(t.Context(), rateLimitKey, maxAttempts, time.Minute)

	assert.Error(t, err)
}

func newRateLimiter(t *testing.T, instance testvalkey.Instance) *valkey.RateLimiter {
	t.Helper()

	return valkey.NewRateLimiter(open(t, instance))
}

func hitTimes(t *testing.T, limiter *valkey.RateLimiter, times int, window time.Duration) {
	t.Helper()

	for range times {
		_, _, err := limiter.Hit(t.Context(), rateLimitKey, maxAttempts, window)
		require.NoError(t, err)
	}
}

func assertAllowed(t *testing.T, limiter *valkey.RateLimiter, key string, window time.Duration) {
	t.Helper()

	allowed, _, err := limiter.Hit(t.Context(), key, maxAttempts, window)
	require.NoError(t, err)
	assert.True(t, allowed, "the attempt was refused")
}

func assertRefused(t *testing.T, limiter *valkey.RateLimiter, window time.Duration) time.Duration {
	t.Helper()

	allowed, retryAfter, err := limiter.Hit(t.Context(), rateLimitKey, maxAttempts, window)
	require.NoError(t, err)
	assert.False(t, allowed, "the attempt was allowed")

	return retryAfter
}
