//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

func TestSessionsUseTheSchemaAndTheTimeouts(t *testing.T) {
	t.Parallel()

	config := testdb.New(t).PostgresConfig()
	config.Schema = "Tenant One"
	config.StatementTimeout = 1500 * time.Millisecond
	config.IdleInTransactionTimeout = time.Minute

	pool := open(t, config)
	require.NoError(t, pool.EnsureSchema(t.Context()))

	assert.Equal(t, "Tenant One", queryString(t, pool, "SELECT current_schema()"))
	assert.Equal(t, "1500ms", queryString(t, pool, "SHOW statement_timeout"))
	assert.Equal(t, "1min", queryString(t, pool, "SHOW idle_in_transaction_session_timeout"))
}

func TestZeroTimeoutsLeaveTheServerDefaults(t *testing.T) {
	t.Parallel()

	pool := open(t, testdb.New(t).PostgresConfig())

	assert.Equal(t, "0", queryString(t, pool, "SHOW statement_timeout"))
	assert.Equal(t, "0", queryString(t, pool, "SHOW idle_in_transaction_session_timeout"))
}

func TestPingReachesTheDatabase(t *testing.T) {
	t.Parallel()

	pool := open(t, testdb.New(t).PostgresConfig())

	assert.NoError(t, pool.Ping(t.Context()))
}

func TestPingFailsWhileTheDatabaseIsUnreachable(t *testing.T) {
	t.Parallel()

	config := testdb.Unreachable(t).PostgresConfig()
	config.ConnectTimeout = time.Second

	assert.Error(t, open(t, config).Ping(t.Context()))
}

func open(t *testing.T, config postgres.Config) *postgres.Pool {
	t.Helper()

	pool, err := postgres.Open(t.Context(), config)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}

func queryString(t *testing.T, pool *postgres.Pool, query string) string {
	t.Helper()

	db := pool.OpenSQLDB()

	t.Cleanup(func() { _ = db.Close() })

	var value string

	require.NoError(t, db.QueryRowContext(t.Context(), query).Scan(&value))

	return value
}
