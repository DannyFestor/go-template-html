package bootstrap

import (
	"time"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	authjob "github.com/DannyFestor/go-template-html/internal/job/auth"
	"github.com/DannyFestor/go-template-html/internal/worker"
)

const pruneExpiredRememberTokensInterval = 24 * time.Hour

func registerAuthJobs(jobs *worker.Registry, pool *postgres.Pool) {
	prune := auth.NewPruneExpiredRememberTokens(postgres.NewRememberTokenStore(pool), clock.System{})

	worker.RegisterPeriodic(
		jobs,
		authjob.NewPruneExpiredRememberTokensHandler(prune),
		worker.DefaultSettings(),
		pruneExpiredRememberTokensInterval,
	)
}
