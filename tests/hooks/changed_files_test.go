package hooks_test

import "testing"

const changedFilesScript = "changed-files.sh"

func TestChangedFilesListsOnlyTheBranchChangesSinceTheMergeBase(t *testing.T) {
	t.Parallel()

	upstream := newRepo(t)
	clone := cloneRepo(t, upstream)
	git(t, clone, "switch", "--quiet", "--create", "feature")
	commitFile(t, clone, "internal/feature/feature.go")
	commitFile(t, upstream, "internal/later/later.go")
	git(t, clone, "fetch", "--quiet")

	run := runScript(t, clone, "", changedFilesScript)

	requirePassed(t, run)
	requireLines(t, run.output, "internal/feature/feature.go")
}

func TestChangedFilesListsEveryTrackedFileWithoutARemoteMain(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	commitFile(t, repo, "cmd/web/main.go")

	run := runScript(t, repo, "", changedFilesScript)

	requirePassed(t, run)
	requireLines(t, run.output, "README.md", "cmd/web/main.go")
}
