package queue

import (
	"context"
	"time"
)

//go:generate go tool mockgen -source=failed_job_deleter.go -destination=mock_failed_job_deleter_test.go -package=queue -mock_names=failedJobDeleter=MockFailedJobDeleter

type failedJobDeleter interface {
	DeleteFailedBefore(ctx context.Context, failedBefore time.Time) error
}
