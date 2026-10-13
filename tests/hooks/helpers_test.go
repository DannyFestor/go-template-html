package hooks_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	scriptsDir   = "../../scripts"
	protectedRef = "refs/heads/main"
)

// Hooks export GIT_DIR and friends, which would point every git call at the outer repository. The global and
// system configs are replaced too, so a developer's signing or hooksPath settings cannot leak in.
func isolatedEnv() []string {
	env := []string{"GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_CONFIG_NOSYSTEM=1"}

	for _, variable := range os.Environ() {
		if !strings.HasPrefix(variable, "GIT_") {
			env = append(env, variable)
		}
	}

	return env
}

func newRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	git(t, dir, "init", "--quiet", "--initial-branch=main")
	configureAuthor(t, dir)
	commitFile(t, dir, "README.md")

	return dir
}

func cloneRepo(t *testing.T, upstream string) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "clone")
	git(t, upstream, "clone", "--quiet", upstream, dir)
	configureAuthor(t, dir)

	return dir
}

func configureAuthor(t *testing.T, dir string) {
	t.Helper()

	git(t, dir, "config", "user.name", "Hook Test")
	git(t, dir, "config", "user.email", "hook-test@example.com")
}

func commitFile(t *testing.T, dir, name string) {
	t.Helper()

	commitContent(t, dir, name, t.Name()+"\n")
}

func commitContent(t *testing.T, dir, name, content string) {
	t.Helper()

	writeFile(t, dir, name, content)
	git(t, dir, "add", name)
	git(t, dir, "commit", "--quiet", "--message", "add "+name)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)

	err := os.MkdirAll(filepath.Dir(path), 0o750)
	if err != nil {
		t.Fatalf("create directory for %s: %v", name, err)
	}

	err = os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func isolatedCommand(t *testing.T, dir, name string, args ...string) *exec.Cmd {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), name, args...) //nolint:gosec // The tests choose every argument.
	cmd.Dir = dir
	cmd.Env = isolatedEnv()

	return cmd
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()

	output, err := isolatedCommand(t, dir, "git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

type scriptRun struct {
	output string
	err    error
}

func runScript(t *testing.T, dir, stdin, script string, args ...string) scriptRun {
	t.Helper()

	path, err := filepath.Abs(filepath.Join(scriptsDir, script))
	if err != nil {
		t.Fatalf("resolve %s: %v", script, err)
	}

	cmd := isolatedCommand(t, dir, path, args...)
	cmd.Stdin = strings.NewReader(stdin)

	output, err := cmd.CombinedOutput()

	return scriptRun{output: string(output), err: err}
}

func requireRefused(t *testing.T, run scriptRun, wantMessage string) {
	t.Helper()

	if run.err == nil {
		t.Fatalf("script passed, want it refused\n%s", run.output)
	}

	if !strings.Contains(run.output, wantMessage) {
		t.Errorf("output = %q, want it to contain %q", run.output, wantMessage)
	}
}

func requirePassed(t *testing.T, run scriptRun) {
	t.Helper()

	if run.err != nil {
		t.Fatalf("script refused: %v\n%s", run.err, run.output)
	}
}

func requireLines(t *testing.T, output string, want ...string) {
	t.Helper()

	got := strings.Fields(output)
	if !slices.Equal(got, want) {
		t.Errorf("lines = %q, want %q", got, want)
	}
}
