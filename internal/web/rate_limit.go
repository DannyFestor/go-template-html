package web

import (
	"context"
	"time"
)

type rateLimiter interface {
	Hit(ctx context.Context, key string, maxAttempts int, window time.Duration) (bool, time.Duration, error)
}

type Limit struct {
	Name        string
	MaxAttempts int
	Window      time.Duration
}

type RateLimit struct {
	limiter rateLimiter
	limit   Limit
}

func NewRateLimit(limiter rateLimiter, limit Limit) *RateLimit {
	return &RateLimit{limiter: limiter, limit: limit}
}

func (r *RateLimit) Hit(ctx context.Context, key string) error {
	allowed, retryAfter, err := r.limiter.Hit(ctx, r.limit.Name+":"+key, r.limit.MaxAttempts, r.limit.Window)
	if err != nil {
		return ErrServiceUnavailable(err)
	}

	if !allowed {
		return ErrTooManyRequests(retryAfter)
	}

	return nil
}
