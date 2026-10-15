package hooks_test

import "testing"

const (
	failOnChangedMigrationsScript = "fail-on-changed-migrations.sh"
	appliedMigration              = "migrations/00001_create_users.sql"
)

func repoWithAppliedMigration(t *testing.T) string {
	t.Helper()

	repo := newRepo(t)
	commitContent(t, repo, appliedMigration, "-- +goose Up\nCREATE TABLE users (id int);\n")
	git(t, repo, "switch", "--quiet", "--create", "feature")

	return repo
}

func TestFailOnChangedMigrationsRefusesAModifiedMigration(t *testing.T) {
	t.Parallel()

	repo := repoWithAppliedMigration(t)
	commitContent(t, repo, appliedMigration, "-- +goose Up\nCREATE TABLE users (id bigint);\n")

	run := runScript(t, repo, "", failOnChangedMigrationsScript, "main", "migrations")

	requireRefused(t, run, appliedMigration)
}

func TestFailOnChangedMigrationsRefusesADeletedMigration(t *testing.T) {
	t.Parallel()

	repo := repoWithAppliedMigration(t)
	git(t, repo, "rm", "--quiet", appliedMigration)
	git(t, repo, "commit", "--quiet", "--message", "delete the migration")

	run := runScript(t, repo, "", failOnChangedMigrationsScript, "main", "migrations")

	requireRefused(t, run, appliedMigration)
}

func TestFailOnChangedMigrationsPassesANewMigrationAndItsSum(t *testing.T) {
	t.Parallel()

	repo := repoWithAppliedMigration(t)
	commitContent(t, repo, "migrations/00002_create_posts.sql", "-- +goose Up\nCREATE TABLE posts (id int);\n")
	commitContent(t, repo, "migrations/migrations.sum", "h1:changed\n")

	run := runScript(t, repo, "", failOnChangedMigrationsScript, "main", "migrations")

	requirePassed(t, run)
}

func TestFailOnChangedMigrationsFailsWithoutTheBaseRef(t *testing.T) {
	t.Parallel()

	repo := repoWithAppliedMigration(t)

	run := runScript(t, repo, "", failOnChangedMigrationsScript, "origin/main", "migrations")

	requireRefused(t, run, "origin/main")
}
