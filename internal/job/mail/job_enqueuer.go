package mail

import "context"

//go:generate go tool mockgen -source=job_enqueuer.go -destination=mock_job_enqueuer_test.go -package=mail -mock_names=jobEnqueuer=MockJobEnqueuer
type jobEnqueuer interface {
	Enqueue(ctx context.Context, job job) error
}
