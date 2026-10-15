//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

var errCallback = errors.New("callback failed")

func TestATransactionCommitsTheWritesOfItsCallback(t *testing.T) {
	t.Parallel()

	database, pool := markerDatabase(t)

	err := postgres.NewTransactor(pool).WithinTransaction(t.Context(), func(ctx context.Context) error {
		insertMarker(ctx, t, pool, "outer")

		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"outer"}, markers(t, database))
}

func TestANestedTransactionRollsBackWithTheOuterOne(t *testing.T) {
	t.Parallel()

	database, pool := markerDatabase(t)
	transactor := postgres.NewTransactor(pool)

	err := transactor.WithinTransaction(t.Context(), func(ctx context.Context) error {
		require.NoError(t, transactor.WithinTransaction(ctx, func(ctx context.Context) error {
			insertMarker(ctx, t, pool, "inner")

			return nil
		}))

		return errCallback
	})

	require.ErrorIs(t, err, errCallback)
	assert.Empty(t, markers(t, database))
}

func TestAnErrorInANestedTransactionRollsBackBoth(t *testing.T) {
	t.Parallel()

	database, pool := markerDatabase(t)
	transactor := postgres.NewTransactor(pool)

	err := transactor.WithinTransaction(t.Context(), func(ctx context.Context) error {
		insertMarker(ctx, t, pool, "outer")

		err := transactor.WithinTransaction(ctx, func(ctx context.Context) error {
			insertMarker(ctx, t, pool, "inner")

			return errCallback
		})

		return fmt.Errorf("outer: %w", err)
	})

	require.ErrorIs(t, err, errCallback)
	assert.Empty(t, markers(t, database))
}

func markerDatabase(t *testing.T) (testdb.Database, *postgres.Pool) {
	t.Helper()

	database := testdb.New(t)

	_, err := connect(t, database).Exec(t.Context(), "CREATE TABLE marker (name text NOT NULL)")
	require.NoError(t, err)

	return database, open(t, database.PostgresConfig())
}

func insertMarker(ctx context.Context, t *testing.T, pool *postgres.Pool, name string) {
	t.Helper()

	require.NoError(t, postgres.Exec(ctx, pool, "INSERT INTO marker (name) VALUES ($1)", name))
}

func markers(t *testing.T, database testdb.Database) []string {
	t.Helper()

	rows, err := connect(t, database).Query(t.Context(), "SELECT name FROM marker ORDER BY name")
	require.NoError(t, err)

	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	return names
}

func connect(t *testing.T, database testdb.Database) *pgx.Conn {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), database.ConnString())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(context.WithoutCancel(t.Context())) })

	return conn
}
