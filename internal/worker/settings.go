package worker

import "time"

const (
	defaultLease   = 5 * time.Minute
	defaultRetries = 10
)

type Settings struct {
	Lease   time.Duration
	Retries int
	Backoff func(attempt int) time.Duration
}

func DefaultSettings() Settings {
	return Settings{Lease: defaultLease, Retries: defaultRetries, Backoff: PolynomialBackoff}
}

func (s Settings) lastAttempt() int {
	return s.Retries + 1
}
