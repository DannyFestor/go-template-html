package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"text/template"

	// The binary is built at run time, which the go test cache cannot see. Importing the adapter, and through it
	// every package the binary holds, makes a source change invalidate cached results.
	_ "github.com/DannyFestor/go-template-html/tools/agenthook/internal/claude"
)

//nolint:gochecknoglobals // TestMain builds the binary once for every test.
var binaryPath string

func TestMain(m *testing.M) {
	os.Exit(runWithBinary(m))
}

func runWithBinary(m *testing.M) int {
	dir, err := os.MkdirTemp("", "agenthook-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}
	defer func() { _ = os.RemoveAll(dir) }()

	binaryPath = filepath.Join(dir, "agenthook")

	output, err := isolatedCommand(context.Background(), ".", "go", "build", "-o", binaryPath, ".").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build agenthook: %v\n%s", err, output)

		return 1
	}

	return m.Run()
}

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

func isolatedCommand(ctx context.Context, dir, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // The tests choose every argument.
	cmd.Dir = dir
	cmd.Env = isolatedEnv()

	return cmd
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()

	output, err := isolatedCommand(t.Context(), dir, "git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

// newRepo returns a checkout on main whose Makefile only logs the targets the hook asks it to run.
func newRepo(t *testing.T) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "repo")
	git(t, t.TempDir(), "init", "--quiet", "--initial-branch=main", dir)
	git(t, dir, "config", "user.name", "Hook Test")
	git(t, dir, "config", "user.email", "hook-test@example.com")

	stub, err := os.ReadFile(filepath.Join("testdata", "Makefile.stub"))
	if err != nil {
		t.Fatalf("read the stub Makefile: %v", err)
	}

	writeFile(t, dir, "Makefile", string(stub))
	writeFile(t, dir, ".gitignore", "make.log\noutput-*\nfail-*\n")
	commitAll(t, dir)

	return dir
}

func addWorktree(t *testing.T, repo, branch string) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), branch)
	git(t, repo, "worktree", "add", "--quiet", "-b", branch, dir)

	return dir
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)

	err := os.MkdirAll(filepath.Dir(path), 0o750)
	if err != nil {
		t.Fatalf("create directory for %s: %v", name, err)
	}

	err = os.WriteFile(path, []byte(content), 0o600) //nolint:gosec // The tests choose every path.
	if err != nil {
		t.Fatalf("write %s: %v", name, err)
	}

	return path
}

func commitFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := writeFile(t, dir, name, content)
	commitAll(t, dir)

	return path
}

func commitAll(t *testing.T, dir string) {
	t.Helper()

	git(t, dir, "add", ".")
	git(t, dir, "commit", "--quiet", "--message", "commit "+t.Name())
}

// hookCall fills the placeholders of a recorded hook JSON fixture in testdata/.
type hookCall struct {
	SessionID string
	Cwd       string
	FilePath  string
	OldString string
	NewString string
	Content   string
	Command   string
}

type hookResult struct {
	stdout   string
	stderr   string
	exitCode int
	response hookResponse
}

type hookResponse struct {
	HookSpecificOutput struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason"`
	} `json:"hookSpecificOutput"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func runHook(t *testing.T, event, fixture string, call hookCall) hookResult {
	t.Helper()

	cmd := isolatedCommand(t.Context(), call.Cwd, binaryPath, event)
	cmd.Stdin = strings.NewReader(renderFixture(t, fixture, call))

	var stdout, stderr bytes.Buffer

	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	var result hookResult

	err := cmd.Run()

	if exitErr, exited := errors.AsType[*exec.ExitError](err); exited {
		result.exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("run agenthook %s: %v", event, err)
	}

	result.stdout, result.stderr = stdout.String(), stderr.String()
	decodeResponse(t, &result)

	return result
}

func renderFixture(t *testing.T, fixture string, call hookCall) string {
	t.Helper()

	functions := template.FuncMap{"json": func(value string) (string, error) {
		encoded, err := json.Marshal(value)

		return string(encoded), err
	}}

	name := fixture + ".json.tmpl"

	parsed, err := template.New(name).Funcs(functions).ParseFiles(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("parse fixture %s: %v", fixture, err)
	}

	var rendered strings.Builder

	err = parsed.ExecuteTemplate(&rendered, name, call)
	if err != nil {
		t.Fatalf("render fixture %s: %v", fixture, err)
	}

	return rendered.String()
}

func decodeResponse(t *testing.T, result *hookResult) {
	t.Helper()

	if result.stdout == "" {
		return
	}

	err := json.Unmarshal([]byte(result.stdout), &result.response)
	if err != nil {
		t.Fatalf("decode hook output %q: %v", result.stdout, err)
	}
}

func requirePermission(t *testing.T, result hookResult, wantDecision, wantReason string) {
	t.Helper()

	got := result.response.HookSpecificOutput
	if result.exitCode != 0 || got.PermissionDecision != wantDecision {
		t.Fatalf("exit %d, decision %q, want exit 0 and %q\nstdout: %s\nstderr: %s",
			result.exitCode, got.PermissionDecision, wantDecision, result.stdout, result.stderr)
	}

	if !strings.Contains(got.PermissionDecisionReason, wantReason) {
		t.Errorf("reason = %q, want it to contain %q", got.PermissionDecisionReason, wantReason)
	}
}

func requireBlocked(t *testing.T, result hookResult, wantReason string) {
	t.Helper()

	if result.exitCode != 0 || result.response.Decision != "block" {
		t.Fatalf("exit %d, decision %q, want exit 0 and block\nstdout: %s\nstderr: %s",
			result.exitCode, result.response.Decision, result.stdout, result.stderr)
	}

	if !strings.Contains(result.response.Reason, wantReason) {
		t.Errorf("reason = %q, want it to contain %q", result.response.Reason, wantReason)
	}
}

func requirePassed(t *testing.T, result hookResult) {
	t.Helper()

	if result.exitCode != 0 || result.stdout != "" {
		t.Fatalf(
			"exit %d, stdout %q, want exit 0 and no output\nstderr: %s",
			result.exitCode,
			result.stdout,
			result.stderr,
		)
	}
}

// makeLog lists the targets the stub Makefile in dir ran, each with its GO_MODULE and FILES.
func makeLog(t *testing.T, dir string) []string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(dir, "make.log")) //nolint:gosec // The tests choose every path.
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		t.Fatalf("read make.log: %v", err)
	}

	return strings.Split(strings.TrimSpace(string(content)), "\n")
}

func requireMakeLog(t *testing.T, dir string, want ...string) {
	t.Helper()

	got := makeLog(t, dir)
	if !slices.Equal(got, want) {
		t.Errorf("make ran\n  %s\nwant\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
}

func clearMakeLog(t *testing.T, dir string) {
	t.Helper()

	err := os.Remove(filepath.Join(dir, "make.log"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("clear make.log: %v", err)
	}
}
