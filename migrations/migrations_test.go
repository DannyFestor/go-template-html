package migrations_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/migrationsum"
	"github.com/DannyFestor/go-template-html/migrations"
)

const downSection = "-- +goose Down"

func TestTheSumMatchesTheMigrations(t *testing.T) {
	t.Parallel()

	assert.NoError(t, migrationsum.Verify(migrations.FS))
}

func TestNoMigrationHasADownSection(t *testing.T) {
	t.Parallel()

	names, err := fs.Glob(migrations.FS, "*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, names)

	for _, name := range names {
		content, err := fs.ReadFile(migrations.FS, name)
		require.NoError(t, err)
		assert.NotContains(t, strings.ToLower(string(content)), strings.ToLower(downSection), name)
	}
}
