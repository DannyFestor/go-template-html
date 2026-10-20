package bootstrap_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

func TestTheWorkerRefusesAPoolSmallerThanItsConcurrencyPlusTwo(t *testing.T) {
	t.Parallel()

	config, err := bootstrap.LoadConfig(missingDotenv(t), append(requiredValues(),
		"QUEUE_CONCURRENCY=10",
		"WORKER_DB_MAX_CONNS=11",
	))
	require.NoError(t, err)

	_, err = bootstrap.NewWorkerApp(t.Context(), config, slog.New(slog.DiscardHandler), nil)

	assert.ErrorContains(t, err, "WORKER_DB_MAX_CONNS 11 is below QUEUE_CONCURRENCY 10 + 2")
}
