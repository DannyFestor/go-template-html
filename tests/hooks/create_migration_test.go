package hooks_test

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	createMigrationScript = "create-migration.sh"
	upOnlyStub            = "-- +goose Up\n"
)

func TestCreateMigrationStartsTheSequenceAtOne(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "migrations.sum", "h1:x\n")
	writeFile(t, dir, "embed.go", "package migrations\n")

	run := runScript(t, dir, "", createMigrationScript, dir, "create_users")

	requirePassed(t, run)
	requireContent(t, dir, "00001_create_users.sql", upOnlyStub)
}

func TestCreateMigrationTakesTheNumberAfterTheHighest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "00009_create_users.sql", upOnlyStub)
	writeFile(t, dir, "00099_create_posts.sql", upOnlyStub)

	run := runScript(t, dir, "", createMigrationScript, dir, "add_tags")

	requirePassed(t, run)
	requireContent(t, dir, "00100_add_tags.sql", upOnlyStub)
}

func TestCreateMigrationPrintsThePathItWrote(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	run := runScript(t, dir, "", createMigrationScript, dir, "create_users")

	requireOutput(t, run, filepath.Join(dir, "00001_create_users.sql"))
}

func TestCreateMigrationRefusesAMissingOrInvalidName(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"", "Create Users", "create-users", "../escape"} {
		dir := t.TempDir()

		run := runScript(t, dir, "", createMigrationScript, dir, name)

		requireRefused(t, run, "NAME")
		requireEmptyDir(t, dir)
	}
}

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("list %s: %v", dir, err)
	}

	if len(entries) > 0 {
		t.Errorf("%s holds %d entries, want none", dir, len(entries))
	}
}
