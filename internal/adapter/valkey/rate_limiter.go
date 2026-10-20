package valkey

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const rateLimitArea = "ratelimit:"

type RateLimiter struct {
	client *Client
}

func NewRateLimiter(client *Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (l *RateLimiter) Hit(
	ctx context.Context,
	key string,
	maxAttempts int,
	window time.Duration,
) (bool, time.Duration, error) {
	attempts, remaining, err := l.count(ctx, l.client.key(rateLimitArea, key), window)
	if err != nil {
		return false, 0, fmt.Errorf("count the attempt: %w", err)
	}

	return attempts <= int64(maxAttempts), remaining, nil
}

func (l *RateLimiter) Clear(ctx context.Context, key string) error {
	err := l.client.redis.Del(ctx, l.client.key(rateLimitArea, key)).Err()
	if err != nil {
		return fmt.Errorf("clear the attempts: %w", err)
	}

	return nil
}

func (l *RateLimiter) count(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	var (
		attempts  *redis.IntCmd
		remaining *redis.DurationCmd
	)

	_, err := l.client.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SetNX(ctx, key, 0, window)
		attempts = pipe.Incr(ctx, key)
		remaining = pipe.PTTL(ctx, key)

		return nil
	})
	if err != nil {
		return 0, 0, err //nolint:wrapcheck // Hit wraps it.
	}

	return attempts.Val(), remaining.Val(), nil
}
