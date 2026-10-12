package hooks_test

import "testing"

const (
	refusePushScript = "refuse-push-to-ref.sh"
	localSHA         = "1111111111111111111111111111111111111111"
	remoteSHA        = "2222222222222222222222222222222222222222"
)

func prePushStdinLine(remoteRef string) string {
	return "refs/heads/feature " + localSHA + " " + remoteRef + " " + remoteSHA + "\n"
}

func TestRefusePushToRefRefusesThePushedProtectedRef(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stdin := prePushStdinLine("refs/heads/feature") + prePushStdinLine(protectedRef)

	run := runScript(t, dir, stdin, refusePushScript, protectedRef)

	requireRefused(t, run, "pushes to main are refused")
}

func TestRefusePushToRefPassesAPushedBranch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	run := runScript(t, dir, prePushStdinLine("refs/heads/feature"), refusePushScript, protectedRef)

	requirePassed(t, run)
}
