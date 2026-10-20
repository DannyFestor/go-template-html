package hooks_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	worktreeEnvScript = "worktree-env.sh"
	defaultDBName     = "go_template_html"
	mainEnv           = "APP_NAME=Go Template\nDB_PORT=5434\n"
)

var testDBNameLine = regexp.MustCompile(`(?m)^TEST_DB_NAME=(\w+)_wt_[0-9a-f]{6}$`)

func TestWorktreeEnvCopiesTheMainEnvAndAddsATestDatabaseName(t *testing.T) {
	t.Parallel()

	main := newRepo(t)
	writeFile(t, main, ".env", mainEnv)
	worktree := addWorktree(t, main, "feature")

	run := runScript(t, worktree, "", worktreeEnvScript, defaultDBName)

	requirePassed(t, run)
	env := readEnv(t, worktree)
	requireEnvStartsWith(t, env, mainEnv)
	requireTestDBName(t, env, defaultDBName)
}

func TestWorktreeEnvBasesTheTestDatabaseNameOnTheMainEnvsDBName(t *testing.T) {
	t.Parallel()

	main := newRepo(t)
	writeFile(t, main, ".env", "DB_NAME=shop\n")
	worktree := addWorktree(t, main, "feature")

	run := runScript(t, worktree, "", worktreeEnvScript, defaultDBName)

	requirePassed(t, run)
	requireTestDBName(t, readEnv(t, worktree), "shop")
}

func TestWorktreeEnvEndsAMainEnvWithoutAFinalNewlineBeforeItsLine(t *testing.T) {
	t.Parallel()

	main := newRepo(t)
	writeFile(t, main, ".env", "DB_PORT=5434")
	worktree := addWorktree(t, main, "feature")

	run := runScript(t, worktree, "", worktreeEnvScript, defaultDBName)

	requirePassed(t, run)
	env := readEnv(t, worktree)
	requireEnvStartsWith(t, env, "DB_PORT=5434\n")
	requireTestDBName(t, env, defaultDBName)
}

func TestWorktreeEnvGivesEachWorktreeItsOwnTestDatabaseName(t *testing.T) {
	t.Parallel()

	main := newRepo(t)
	writeFile(t, main, ".env", mainEnv)
	first := addWorktree(t, main, "first")
	second := addWorktree(t, main, "second")

	requirePassed(t, runScript(t, first, "", worktreeEnvScript, defaultDBName))
	requirePassed(t, runScript(t, second, "", worktreeEnvScript, defaultDBName))

	if testDBNameLine.FindString(readEnv(t, first)) == testDBNameLine.FindString(readEnv(t, second)) {
		t.Error("both worktrees got the same TEST_DB_NAME")
	}
}

func TestWorktreeEnvRefusesTheMainCheckout(t *testing.T) {
	t.Parallel()

	main := newRepo(t)

	run := runScript(t, main, "", worktreeEnvScript, defaultDBName)

	requireRefused(t, run, "linked worktree")
	requireNoFile(t, main, ".env")
}

func TestWorktreeEnvKeepsAnEnvTheWorktreeAlreadyHas(t *testing.T) {
	t.Parallel()

	main := newRepo(t)
	writeFile(t, main, ".env", mainEnv)
	worktree := addWorktree(t, main, "feature")
	writeFile(t, worktree, ".env", "DB_PORT=6000\n")

	run := runScript(t, worktree, "", worktreeEnvScript, defaultDBName)

	requireRefused(t, run, "already exists")
	requireContent(t, worktree, ".env", "DB_PORT=6000\n")
}

func TestWorktreeEnvRefusesWhenTheMainCheckoutHasNoEnv(t *testing.T) {
	t.Parallel()

	main := newRepo(t)
	worktree := addWorktree(t, main, "feature")

	run := runScript(t, worktree, "", worktreeEnvScript, defaultDBName)

	requireRefused(t, run, ".env.example")
	requireNoFile(t, worktree, ".env")
}

func addWorktree(t *testing.T, main, name string) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), name)
	git(t, main, "worktree", "add", "--quiet", "-b", name, dir)

	return dir
}

func readEnv(t *testing.T, dir string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(dir, ".env")) //nolint:gosec // The tests choose every path.
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}

	return string(content)
}

func requireEnvStartsWith(t *testing.T, env, want string) {
	t.Helper()

	if !strings.HasPrefix(env, want) {
		t.Errorf(".env = %q, want it to start with %q", env, want)
	}
}

func requireTestDBName(t *testing.T, env, wantBase string) {
	t.Helper()

	match := testDBNameLine.FindStringSubmatch(env)
	if match == nil {
		t.Fatalf(".env = %q, want a TEST_DB_NAME=<name>_wt_<6 hex digits> line", env)
	}

	if match[1] != wantBase {
		t.Errorf("TEST_DB_NAME is based on %q, want %q", match[1], wantBase)
	}
}

func requireNoFile(t *testing.T, dir, name string) {
	t.Helper()

	_, err := os.Stat(filepath.Join(dir, name))
	if !os.IsNotExist(err) {
		t.Errorf("%s exists, want it absent", name)
	}
}
