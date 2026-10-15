//go:build integration

package goose_test

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/goose"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

const (
	createUsers = "00001_create_users.sql"
	createPosts = "00002_create_posts.sql"
	schema      = "migrator_test"
)

func TestUpAppliesThePendingMigrationsInTheSchema(t *testing.T) {
	t.Parallel()

	pool := openPool(t)

	applied, err := newMigrator(t, pool, usersAndPosts()).Up(t.Context())

	require.NoError(t, err)
	assert.Equal(t, []string{createUsers, createPosts}, applied)
	assert.True(t, tableExists(t, pool, schema+".posts"))
	assert.True(t, tableExists(t, pool, schema+".goose_db_version"))
}

func TestUpAppliesNothingTheSecondTime(t *testing.T) {
	t.Parallel()

	pool := openPool(t)
	migrator := newMigrator(t, pool, usersAndPosts())

	_, err := migrator.Up(t.Context())
	require.NoError(t, err)

	applied, err := migrator.Up(t.Context())

	require.NoError(t, err)
	assert.Empty(t, applied)
}

func TestStatusTellsAppliedFromPendingMigrations(t *testing.T) {
	t.Parallel()

	pool := openPool(t)
	_, err := newMigrator(t, pool, fstest.MapFS{createUsers: migration("CREATE TABLE users (id int);")}).Up(t.Context())
	require.NoError(t, err)

	statuses, err := newMigrator(t, pool, usersAndPosts()).Status(t.Context())

	require.NoError(t, err)
	require.Len(t, statuses, 2)
	assert.Equal(t, createUsers, statuses[0].Name)
	assert.True(t, statuses[0].Applied)
	assert.WithinDuration(t, time.Now(), statuses[0].AppliedAt, time.Minute)
	assert.Equal(t, goose.Status{Name: createPosts, Applied: false, AppliedAt: time.Time{}}, statuses[1])
}

func usersAndPosts() fstest.MapFS {
	return fstest.MapFS{
		createUsers: migration("CREATE TABLE users (id int);"),
		createPosts: migration("CREATE TABLE posts (id int);"),
	}
}

func migration(statement string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte("-- +goose Up\n" + statement + "\n"), Mode: 0, ModTime: time.Time{}, Sys: nil}
}

func openPool(t *testing.T) *postgres.Pool {
	t.Helper()

	config := testdb.New(t).PostgresConfig()
	config.Schema = schema

	pool, err := postgres.Open(t.Context(), config)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pool.EnsureSchema(t.Context()))

	return pool
}

func newMigrator(t *testing.T, pool *postgres.Pool, migrations fstest.MapFS) *goose.Migrator {
	t.Helper()

	db := pool.OpenSQLDB()

	t.Cleanup(func() { _ = db.Close() })

	migrator, err := goose.New(db, migrations)
	require.NoError(t, err)

	return migrator
}

func tableExists(t *testing.T, pool *postgres.Pool, table string) bool {
	t.Helper()

	var exists bool

	db := pool.OpenSQLDB()

	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT to_regclass($1) IS NOT NULL", table).Scan(&exists))

	return exists
}
