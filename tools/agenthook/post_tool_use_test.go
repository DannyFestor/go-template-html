package main_test

import "testing"

func TestEditingAGoFileInALinkedWorktreeRunsTheChecksOfThatWorktreesModule(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	worktree := addWorktree(t, repo, "feature")
	path := writeFile(t, worktree, "internal/a/a.go", "package a\n")

	requirePassed(t, runHook(t, "post-tool-use", "post_tool_use_edit", editCall(repo, path, "package b", "package a")))

	requireMakeLog(t, worktree,
		"fmt GO_MODULE=. FILES=internal/a/a.go",
		"vet GO_MODULE=. FILES=internal/a/a.go",
		"lint-wip GO_MODULE=. FILES=internal/a/a.go",
	)
	requireMakeLog(t, repo)
}

func TestEditingAGoFileInANestedModuleRunsTheChecksInThatModule(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	writeFile(t, repo, "tools/agenthook/go.mod", "module example.com/agenthook\n")
	path := writeFile(t, repo, "tools/agenthook/internal/shell/parse.go", "package shell\n")

	requirePassed(
		t,
		runHook(t, "post-tool-use", "post_tool_use_edit", editCall(repo, path, "package x", "package shell")),
	)

	requireMakeLog(t, repo,
		"fmt GO_MODULE=tools/agenthook FILES=internal/shell/parse.go",
		"vet GO_MODULE=tools/agenthook FILES=internal/shell/parse.go",
		"lint-wip GO_MODULE=tools/agenthook FILES=internal/shell/parse.go",
	)
}

func TestAFailingPostEditCheckIsFedBackToTheAgent(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	writeFile(t, repo, "fail-vet", "internal/a/a.go:3:5: declared and not used: n\n")
	path := writeFile(t, repo, "internal/a/a.go", "package a\n")

	result := runHook(t, "post-tool-use", "post_tool_use_edit", editCall(repo, path, "package b", "package a"))

	requireBlocked(t, result, "declared and not used: n")
	requireBlocked(t, result, "make vet")
	requireMakeLog(t, repo, "fmt GO_MODULE=. FILES=internal/a/a.go", "vet GO_MODULE=. FILES=internal/a/a.go")
}

func TestWritingATypeScriptFileRunsOxfmtAndOxlintOnIt(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	path := writeFile(t, repo, "internal/asset/src/menu.ts", "export {};\n")
	call := writeCall(repo, path, "export {};\n")

	requirePassed(t, runHook(t, "post-tool-use", "post_tool_use_write", call))

	requireMakeLog(t, repo,
		"fmt-js GO_MODULE= FILES=internal/asset/src/menu.ts",
		"lint-js GO_MODULE= FILES=internal/asset/src/menu.ts",
	)
}

func TestEditingAnotherFileRunsNoCheck(t *testing.T) {
	t.Parallel()

	repo := newRepo(t)
	path := writeFile(t, repo, "README.md", "# App\n")

	requirePassed(t, runHook(t, "post-tool-use", "post_tool_use_edit", editCall(repo, path, "# Template", "# App")))

	requireMakeLog(t, repo)
}
