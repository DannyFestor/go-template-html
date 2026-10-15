//go:build integration

package testdb_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

func TestEachCallGetsItsOwnMigratedDatabase(t *testing.T) {
	t.Parallel()

	first, second := testdb.New(t), testdb.New(t)
	require.NotEqual(t, first.Name, second.Name)

	exec(t, first, "CREATE TABLE marker (id int)")

	assert.True(t, tableExists(t, first, "marker"))
	assert.False(t, tableExists(t, second, "marker"))
	assert.True(t, tableExists(t, second, "goose_db_version"))
}

func TestAPassingTestsDatabaseIsDropped(t *testing.T) {
	t.Parallel()

	var name string

	observer := testdb.New(t)

	t.Run("passing", func(t *testing.T) {
		t.Parallel()

		name = testdb.New(t).Name
	})

	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())

		conn, err := pgx.Connect(ctx, observer.ConnString())
		require.NoError(t, err)

		defer func() { _ = conn.Close(ctx) }()

		var exists bool

		err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT FROM pg_database WHERE datname = $1)", name).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, name)
	})
}

func connect(t *testing.T, database testdb.Database) *pgx.Conn {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), database.ConnString())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(context.WithoutCancel(t.Context())) })

	return conn
}

func exec(t *testing.T, database testdb.Database, sql string) {
	t.Helper()

	_, err := connect(t, database).Exec(t.Context(), sql)
	require.NoError(t, err)
}

func tableExists(t *testing.T, database testdb.Database, table string) bool {
	t.Helper()

	var exists bool

	err := connect(t, database).
		QueryRow(t.Context(), "SELECT to_regclass($1) IS NOT NULL", table).
		Scan(&exists)
	require.NoError(t, err)

	return exists
}
