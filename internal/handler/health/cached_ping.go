package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const pingTimeout = 2 * time.Second

// CachedPing holds its lock across the ping, so request rate can never drive more than one ping per TTL.
type CachedPing struct {
	dependency pinger
	ttl        time.Duration
	mutex      sync.Mutex
	expiresAt  time.Time
	result     error
}

func NewCachedPing(dependency pinger, ttl time.Duration) *CachedPing {
	return &CachedPing{dependency: dependency, ttl: ttl, mutex: sync.Mutex{}, expiresAt: time.Time{}, result: nil}
}

func (c *CachedPing) Ping(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if time.Now().Before(c.expiresAt) {
		return c.result
	}

	c.result = c.pingWithTimeout(ctx)
	c.expiresAt = time.Now().Add(c.ttl)

	return c.result
}

// The result is shared with every caller until the TTL runs out, so one caller's cancelled request must not decide it.
func (c *CachedPing) pingWithTimeout(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), pingTimeout)
	defer cancel()

	err := c.dependency.Ping(ctx)
	if err != nil {
		return fmt.Errorf("readiness: %w", err)
	}

	return nil
}
