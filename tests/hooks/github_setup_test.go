package hooks_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	githubSetupScript = "github-setup.sh"
	rulesetFile       = "../../.github/rulesets/main.json"
	mergeSettingsFile = "../../.github/merge-settings.json"

	withAdminRights    = true
	withoutAdminRights = false
)

type fakeGitHub struct {
	bin   string
	state string
}

func newFakeGitHub(t *testing.T) fakeGitHub {
	t.Helper()

	github := fakeGitHub{bin: t.TempDir(), state: t.TempDir()}

	err := os.Symlink(absolutePath(t, "testdata/fake-gh.sh"), filepath.Join(github.bin, "gh"))
	if err != nil {
		t.Fatalf("install the fake gh: %v", err)
	}

	github.writeNewRepo(t, withAdminRights)

	return github
}

func (g fakeGitHub) writeNewRepo(t *testing.T, admin bool) {
	t.Helper()

	writeFile(t, g.state, "repo.json", `{
		"full_name": "owner/repo",
		"permissions": {"admin": `+strconv.FormatBool(admin)+`},
		"allow_merge_commit": true,
		"allow_rebase_merge": true,
		"allow_squash_merge": true,
		"squash_merge_commit_title": "COMMIT_OR_PR_TITLE",
		"squash_merge_commit_message": "COMMIT_MESSAGES",
		"delete_branch_on_merge": false
	}`)
}

func (g fakeGitHub) writeRulesetInEvaluateMode(t *testing.T) {
	t.Helper()

	content, err := os.ReadFile(rulesetFile)
	if err != nil {
		t.Fatalf("read %s: %v", rulesetFile, err)
	}

	var ruleset map[string]any

	err = json.Unmarshal(content, &ruleset)
	if err != nil {
		t.Fatalf("decode %s: %v", rulesetFile, err)
	}

	ruleset["id"] = 1
	ruleset["enforcement"] = "evaluate"

	stored, err := json.Marshal(ruleset)
	if err != nil {
		t.Fatalf("encode the stored ruleset: %v", err)
	}

	writeFile(t, g.state, "rulesets/1.json", string(stored))
}

func (g fakeGitHub) setup(t *testing.T) scriptRun {
	t.Helper()

	return g.setupWithPath(t, g.bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func (g fakeGitHub) setupWithPath(t *testing.T, path string) scriptRun {
	t.Helper()

	cmd := scriptCommand(t, t.TempDir(), githubSetupScript,
		absolutePath(t, rulesetFile), absolutePath(t, mergeSettingsFile))
	cmd.Env = append(cmd.Env, "PATH="+path, "FAKE_GH_STATE="+g.state)

	return runCommand(cmd)
}

func (g fakeGitHub) rulesetCount(t *testing.T) int {
	t.Helper()

	rulesets, err := filepath.Glob(filepath.Join(g.state, "rulesets", "*.json"))
	if err != nil {
		t.Fatalf("list the fake rulesets: %v", err)
	}

	return len(rulesets)
}

func (g fakeGitHub) apiCalls(t *testing.T) []string {
	t.Helper()

	calls, err := os.ReadFile(filepath.Join(g.state, "calls"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		t.Fatalf("read the fake gh calls: %v", err)
	}

	return strings.Split(strings.TrimSpace(string(calls)), "\n")
}

func requireNoWrites(t *testing.T, github fakeGitHub) {
	t.Helper()

	for _, call := range github.apiCalls(t) {
		if !strings.HasPrefix(call, "GET ") {
			t.Errorf("gh api call %q writes, want only reads", call)
		}
	}
}

func requireRulesetCount(t *testing.T, github fakeGitHub, want int) {
	t.Helper()

	if got := github.rulesetCount(t); got != want {
		t.Errorf("rulesets = %d, want %d", got, want)
	}
}

func TestGitHubSetupCreatesTheRuleset(t *testing.T) {
	t.Parallel()

	github := newFakeGitHub(t)

	run := github.setup(t)

	requirePassed(t, run)
	requireOutput(t, run, "created ruleset main")
	requireRulesetCount(t, github, 1)
}

func TestGitHubSetupRerunKeepsOneRulesetAndReportsItUpToDate(t *testing.T) {
	t.Parallel()

	github := newFakeGitHub(t)
	requirePassed(t, github.setup(t))

	run := github.setup(t)

	requirePassed(t, run)
	requireOutput(t, run, "ruleset main already up to date")
	requireOutput(t, run, "merge settings already up to date")
	requireRulesetCount(t, github, 1)
}

func TestGitHubSetupReportsEachMergeSettingItChanged(t *testing.T) {
	t.Parallel()

	github := newFakeGitHub(t)

	run := github.setup(t)

	requirePassed(t, run)
	requireOutput(t, run, "allow_merge_commit: true -> false")
	requireOutput(t, run, "allow_rebase_merge: true -> false")
	requireOutput(t, run, "squash_merge_commit_title: COMMIT_OR_PR_TITLE -> PR_TITLE")
	requireOutput(t, run, "squash_merge_commit_message: COMMIT_MESSAGES -> PR_BODY")
	requireOutput(t, run, "delete_branch_on_merge: false -> true")
}

func TestGitHubSetupRefusesAnUnauthenticatedGh(t *testing.T) {
	t.Parallel()

	github := newFakeGitHub(t)
	writeFile(t, github.state, "logged-out", "")

	run := github.setup(t)

	requireRefused(t, run, "gh is not authenticated, run gh auth login first")
	requireNoWrites(t, github)
}

func TestGitHubSetupRefusesAUserWithoutAdminRights(t *testing.T) {
	t.Parallel()

	github := newFakeGitHub(t)
	github.writeNewRepo(t, withoutAdminRights)

	run := github.setup(t)

	requireRefused(t, run, "you lack admin rights on owner/repo")
	requireNoWrites(t, github)
}

func TestGitHubSetupRefusesToRunWithoutGh(t *testing.T) {
	t.Parallel()

	github := newFakeGitHub(t)

	run := github.setupWithPath(t, t.TempDir())

	requireRefused(t, run, "gh is not installed")
}

func TestGitHubSetupUpdatesAnExistingRulesetAndNamesWhatChanged(t *testing.T) {
	t.Parallel()

	github := newFakeGitHub(t)
	github.writeRulesetInEvaluateMode(t)

	run := github.setup(t)

	requirePassed(t, run)
	requireOutput(t, run, "updated ruleset main: enforcement")
	requireRulesetCount(t, github, 1)
}
