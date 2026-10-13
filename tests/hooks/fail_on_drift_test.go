package hooks_test

import "testing"

const failOnDriftScript = "fail-on-drift.sh"

func TestFailOnDriftFailsOnAChangedTrackedFile(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	commitFile(t, repo, "internal/view/home_templ.go")
	writeFile(t, repo, "internal/view/home_templ.go", "regenerated\n")

	run := runScript(t, repo, "", failOnDriftScript)

	requireRefused(t, run, "make generate")
	requireRefused(t, run, "internal/view/home_templ.go")
}

func TestFailOnDriftFailsOnANewlyGeneratedFile(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	writeFile(t, repo, "internal/mock/store.go", "generated\n")

	run := runScript(t, repo, "", failOnDriftScript)

	requireRefused(t, run, "internal/mock/store.go")
}

func TestFailOnDriftPassesACleanTree(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)

	run := runScript(t, repo, "", failOnDriftScript)

	requirePassed(t, run)
}

func TestFailOnDriftIgnoresGitignoredFiles(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	commitContent(t, repo, ".gitignore", "bin/\n")
	writeFile(t, repo, "bin/cog", "binary\n")

	run := runScript(t, repo, "", failOnDriftScript)

	requirePassed(t, run)
}
