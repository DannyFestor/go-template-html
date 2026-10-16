package bootstrap

import "time"

type Queue struct {
	Concurrency     int           `env:"QUEUE_CONCURRENCY"      envDefault:"10"`
	PollInterval    time.Duration `env:"QUEUE_POLL_INTERVAL"    envDefault:"1s"`
	FailedRetention time.Duration `env:"QUEUE_FAILED_RETENTION" envDefault:"168h"`
}
