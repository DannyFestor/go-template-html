package migrationsum_test

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/migrationsum"
)

const (
	createUsers = "-- +goose Up\nCREATE TABLE users (id uuid);\n"
	createPosts = "-- +goose Up\nCREATE TABLE posts (id uuid);\n"
)

func TestComputeListsTheTotalThenEverySQLFile(t *testing.T) {
	t.Parallel()

	sum, err := migrationsum.Compute(fstest.MapFS{
		"00002_create_posts.sql": file(createPosts),
		"00001_create_users.sql": file(createUsers),
		"embed.go":               file("package migrations\n"),
	})

	require.NoError(t, err)
	assert.Equal(t, "h1:HMd8yKDUGApWIzYZ3Ow6jLJ4IqvrCpSyp3Xv4jvO+Ak=\n"+
		"00001_create_users.sql h1:m5f9n8tuITydDTSl5xpQIdW4N5r0vgdnpru0O7QYSQY=\n"+
		"00002_create_posts.sql h1:qmImPdWmnr6N6F0xJe+8rZ4KS4hq8hAxVOTfUZSwlwA=\n", sum.String())
}

func TestVerifyAcceptsAnUpToDateSum(t *testing.T) {
	t.Parallel()

	assert.NoError(t, migrationsum.Verify(withSum(t, fstest.MapFS{"00001_create_users.sql": file(createUsers)})))
}

func TestVerifyNamesATamperedMigration(t *testing.T) {
	t.Parallel()

	fsys := withSum(t, fstest.MapFS{"00001_create_users.sql": file(createUsers)})
	fsys["00001_create_users.sql"] = file(createPosts)

	err := migrationsum.Verify(fsys)

	require.ErrorIs(t, err, migrationsum.ErrOutOfDate)
	assert.ErrorContains(t, err, "00001_create_users.sql")
}

func TestVerifyNamesAnAddedAndARemovedMigration(t *testing.T) {
	t.Parallel()

	fsys := withSum(t, fstest.MapFS{"00001_create_users.sql": file(createUsers)})
	delete(fsys, "00001_create_users.sql")
	fsys["00002_create_posts.sql"] = file(createPosts)

	err := migrationsum.Verify(fsys)

	require.ErrorIs(t, err, migrationsum.ErrOutOfDate)
	require.ErrorContains(t, err, "00001_create_users.sql")
	assert.ErrorContains(t, err, "00002_create_posts.sql")
}

func TestVerifyFailsWithoutASumFile(t *testing.T) {
	t.Parallel()

	err := migrationsum.Verify(fstest.MapFS{"00001_create_users.sql": file(createUsers)})

	assert.ErrorContains(t, err, migrationsum.FileName)
}

func withSum(t *testing.T, fsys fstest.MapFS) fstest.MapFS {
	t.Helper()

	sum, err := migrationsum.Compute(fsys)
	require.NoError(t, err)

	fsys[migrationsum.FileName] = file(sum.String())

	return fsys
}

func file(content string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(content), Mode: 0, ModTime: time.Time{}, Sys: nil}
}
