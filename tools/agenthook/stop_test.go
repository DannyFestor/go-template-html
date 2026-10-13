package main_test

import (
	"fmt"
	"testing"
)

func stopCall(cwd, sessionID string) hookCall {
	return hookCall{
		SessionID: sessionID,
		Cwd:       cwd,
		FilePath:  "",
		OldString: "",
		NewString: "",
		Content:   "",
		Command:   "",
	}
}

func editGoFile(t *testing.T, cwd, dir, name, sessionID string) string {
	t.Helper()

	path := writeFile(t, dir, name, "package x\n")
	call := editCall(cwd, path, "package y", "package x")
	call.SessionID = sessionID
	requirePassed(t, runHook(t, "post-tool-use", "post_tool_use_edit", call))

	return path
}

// archWarning is go-arch-lint's JSON report of one import that breaks .go-arch-lint.yml.
func archWarning(file string) string {
	return fmt.Sprintf(`{"Type": "models.Check", "Payload": {"ExecutionWarnings": [], "ArchHasWarnings": true,
		"ArchWarningsDeps": [{"ComponentName": "domain", "FileRelativePath": %q,
		"ResolvedImportName": "github.com/DannyFestor/go-template-html/internal/handler"}],
		"ArchWarningsNotMatched": [], "ArchWarningsDeepScan": [], "OmittedCount": 0}}`, "/"+file)
}

func TestTheEndOfTurnGateChecksOnlyThisSessionsPackages(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	editGoFile(t, repo, repo, "internal/a/a.go", "session-a")
	editGoFile(t, repo, repo, "internal/b/b.go", "session-b")
	clearMakeLog(t, repo)

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, "session-b")))

	requireMakeLog(t, repo,
		"build-packages GO_MODULE=. FILES=internal/b/b.go",
		"arch-json GO_MODULE= FILES=",
		"lint GO_MODULE=. FILES=internal/b/b.go",
		"test-go GO_MODULE=. FILES=internal/b/b.go",
	)
}

func TestTheEndOfTurnGateChecksEachModuleOnItsOwn(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	writeFile(t, repo, "tools/agenthook/go.mod", "module example.com/agenthook\n")
	editGoFile(t, repo, repo, "tools/agenthook/main.go", session)
	editGoFile(t, repo, repo, "cmd/web/main.go", session)
	clearMakeLog(t, repo)

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))

	requireMakeLog(t, repo,
		"build-packages GO_MODULE=. FILES=cmd/web/main.go",
		"arch-json GO_MODULE= FILES=",
		"lint GO_MODULE=. FILES=cmd/web/main.go",
		"test-go GO_MODULE=. FILES=cmd/web/main.go",
		"build-packages GO_MODULE=tools/agenthook FILES=main.go",
		"lint GO_MODULE=tools/agenthook FILES=main.go",
		"test-go GO_MODULE=tools/agenthook FILES=main.go",
	)
}

func TestArchWarningsOnAnotherSessionsFilesDoNotBlockTheGate(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	writeFile(t, repo, "output-arch-json", archWarning("internal/a/a.go"))
	writeFile(t, repo, "fail-arch-json", "")
	editGoFile(t, repo, repo, "internal/a/a.go", "session-a")
	editGoFile(t, repo, repo, "internal/b/b.go", "session-b")

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, "session-b")))
	requireBlocked(t, runHook(t, "stop", "stop", stopCall(repo, "session-a")), "internal/a/a.go")
}

func TestTheEndOfTurnGateBlocksOnAFailureAndReleasesAfterThreeBlocksInARow(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	writeFile(t, repo, "fail-test-go", "--- FAIL: TestA\n")
	editGoFile(t, repo, repo, "internal/a/a.go", session)

	for range 3 {
		requireBlocked(t, runHook(t, "stop", "stop", stopCall(repo, session)), "--- FAIL: TestA")
	}

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))
	clearMakeLog(t, repo)

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))
	requireMakeLog(t, repo)
}

func TestTheEndOfTurnGateSkipsWhenNothingChangedSinceItsLastPass(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	path := editGoFile(t, repo, repo, "internal/a/a.go", session)
	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))
	clearMakeLog(t, repo)

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))
	requireMakeLog(t, repo)

	writeFile(t, repo, "internal/a/a.go", "package a\n")
	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))

	if len(makeLog(t, repo)) == 0 {
		t.Errorf("the gate skipped after %s changed", path)
	}
}

func TestTheEndOfTurnGateRunsTheJSChecksWhenTypeScriptWasTouched(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	path := writeFile(t, repo, "internal/asset/src/menu.ts", "export {};\n")
	requirePassed(t, runHook(t, "post-tool-use", "post_tool_use_edit", editCall(repo, path, "", "export {};")))
	clearMakeLog(t, repo)

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))

	requireMakeLog(t, repo, "typecheck-js GO_MODULE= FILES=", "test-js GO_MODULE= FILES=")
}

func TestTheEndOfTurnGateCoversEditsInALinkedWorktree(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	worktree := addWorktree(t, repo, "feature")
	editGoFile(t, repo, worktree, "internal/a/a.go", session)
	clearMakeLog(t, worktree)

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))

	requireMakeLog(t, worktree,
		"build-packages GO_MODULE=. FILES=internal/a/a.go",
		"arch-json GO_MODULE= FILES=",
		"lint GO_MODULE=. FILES=internal/a/a.go",
		"test-go GO_MODULE=. FILES=internal/a/a.go",
	)
	requireMakeLog(t, repo)
}

func TestTheEndOfTurnGatePassesWhenNothingWasEdited(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)

	requirePassed(t, runHook(t, "stop", "stop", stopCall(repo, session)))
	requireMakeLog(t, repo)
}
