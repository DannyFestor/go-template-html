package bootstrap

import (
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/uuidv7"
)

func newJobQueue(pool *postgres.Pool) *postgres.JobQueue {
	return postgres.NewJobQueue(pool, uuidv7.NewGenerator(clock.System{}), clock.System{})
}
