package bootstrap

import (
	"errors"
	"fmt"
)

const connsForClaimingAndRescuing = 2

var errWorkerPoolTooSmall = errors.New("the worker's pool is too small")

func checkWorkerPoolSize(config Config) error {
	minimum := config.Queue.Concurrency + connsForClaimingAndRescuing
	if int(config.DB.Worker.MaxConns) < minimum {
		return fmt.Errorf("%w: WORKER_DB_MAX_CONNS %d is below QUEUE_CONCURRENCY %d + %d",
			errWorkerPoolTooSmall, config.DB.Worker.MaxConns, config.Queue.Concurrency, connsForClaimingAndRescuing)
	}

	return nil
}
