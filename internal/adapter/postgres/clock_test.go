//go:build integration

package postgres_test

import (
	"sync"
	"time"
)

type testClock struct {
	mutex sync.Mutex
	now   time.Time
}

func clockAt(now time.Time) *testClock {
	return &testClock{mutex: sync.Mutex{}, now: now}
}

func (c *testClock) Now() time.Time {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.now
}

func (c *testClock) advance(by time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.now = c.now.Add(by)
}
