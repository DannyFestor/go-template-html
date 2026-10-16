package bootstrap

import "time"

type Queue struct {
	Concurrency  int           `env:"QUEUE_CONCURRENCY"   envDefault:"10"`
	PollInterval time.Duration `env:"QUEUE_POLL_INTERVAL" envDefault:"1s"`
}
