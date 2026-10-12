package hooks_test

import "testing"

const refuseCommitScript = "refuse-commit-on-ref.sh"

func TestRefuseCommitOnRefRefusesTheProtectedBranch(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)

	run := runScript(t, repo, "", refuseCommitScript, protectedRef)

	requireRefused(t, run, "branch first")
}

func TestRefuseCommitOnRefPassesAnotherBranch(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	git(t, repo, "switch", "--quiet", "--create", "feature")

	run := runScript(t, repo, "", refuseCommitScript, protectedRef)

	requirePassed(t, run)
}

func TestRefuseCommitOnRefPassesADetachedHead(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	git(t, repo, "switch", "--quiet", "--detach")

	run := runScript(t, repo, "", refuseCommitScript, protectedRef)

	requirePassed(t, run)
}
