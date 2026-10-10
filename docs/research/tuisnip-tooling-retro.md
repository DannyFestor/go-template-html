# TuiSnip tooling retro: git hooks, agent hooks and CI

Research for issue #33. It collects evidence only. Which rules this template's git and agent hooks enforce is decided later, in the "Git and agent hook rules" ticket. Researched 2026-10-11.

TuiSnip (DannyFestor/TuiSnip) is a Go TUI built mostly by Claude Code agents between 2026-09-30 and 2026-10-09. Its guardrails were lefthook git hooks, Claude Code / OpenCode / Antigravity agent hooks under `scripts/agent/`, GitHub Actions, and a `main` ruleset.

## Sources

- **[C <sha>]**: a commit in the local clone `/Users/danny/git/TuiSnip` (HEAD `00b1da1`). Most are squash merges whose message is the PR body.
- **[I#n]**, **[PR#n]**: an issue or PR in DannyFestor/TuiSnip, body and comments read with `gh`. No PR had review comments that mentioned hooks.
- **[T <session> <time>]**: Claude Code session transcripts in `~/.claude/projects/-Users-danny-git-TuiSnip/`: 108 top-level sessions and 131 subagent transcripts, 2026-09-29 to 2026-10-09 (UTC). These record every hook verdict, so they give counts the issues only estimate. The counts come from throwaway Python scripts that weren't committed. They count events, not distinct problems: a rule that fires five times on one bad edit counts five.
- **[ADR-n]**: this repo's `docs/adr/`.
- `/Users/danny/git/TuiSnip.bak` was not used.

Anything marked *Inference* is a reading of the evidence, not something a source says.

## Summary

| Guardrail | Outcome | One-line lesson |
|---|---|---|
| pre-commit: refuse commit on `main` | Added late, kept | Agents did commit on `main` (16 sessions). The hook caught it cheaply, and AGENTS.md wording then cut the retries. |
| pre-commit: `golangci-lint fmt` + `stage_fixed` | Kept | Ran in about 0.4 s. No partial-staging conflict was seen in practice. |
| pre-commit: shellcheck on staged `*.sh` | Fixed once (`-x`), kept | Needed `-x` from day one, because scripts source a shared lib. |
| commit-msg: `cog verify` | Kept | Rarely failed. The failures were a missing `cog` on PATH and a default merge-commit message. |
| pre-push: refuse push to `main` | Kept | Never caught a real push. The `main` ruleset was the real backstop. |
| pre-push: lint changed packages | Kept | Cheap (median 2 s) and scoped by merge-base. No friction seen. |
| pre-push: all unit tests | Kept | Blocked five docs-only pushes on flaky or leaking tests (median 8 s, max 59 s). |
| pre-push: feature/e2e opt-in | Kept | Opt-in only. The only push that set it was in the session that built the hook. |
| Agent: path guard, generator-output deny | Kept | Never fired in 10 days. |
| Agent: path guard, ask on guardrail files | Kept | Constant prompts while the hooks were being built, quiet otherwise. |
| Agent: new Go comment guard | Loosened twice (exemptions) | About two thirds of first denials were resubmitted, which put an approval prompt in front of the user. |
| Agent: command guard, hook bypass | Kept | No real bypass attempt seen. All three denials were false positives. |
| Agent: command guard, push to `main` / force-push | Reworked (#149) | All 8 push-to-`main` denials were worktree false positives. Force-push denials blocked normal branch rebases. |
| Agent: command guard, shell file writes | Kept, patched 4 times | The most frequent deny (about 140), and a 716-line bash parser that kept needing fixes. |
| Agent: post-edit fmt + vet | Kept, reworked for worktrees | The busiest hook (about 1,360 vet failures). Worktree paths broke it. |
| Agent: post-edit lint | Added (#129) | Moved lint findings from end of turn to edit time, because late findings forced redesigns. |
| Agent: Stop gate | Kept, rarely fired | 12 blocks in 5 sessions. It was released after 3 tries 4 times, once because another session's edits failed it. |
| Agent: multi-agent adapters | Kept | Only Claude Code was used. OpenCode and Antigravity were never tested live. |
| CI `ci.yml` + required checks | Kept, grew `test-macos` | Flakes came from the tests, not the CI design. |
| CI `pr-title` | Kept | Caught PR titles that weren't Conventional Commits (2 sessions). |
| CI nightly → weekly; mutation → manual | Loosened twice | Expensive jobs moved off the schedule once nobody read their reports. |

## Git hooks (`lefthook.yml`)

History: created in [C 5dfd14b] (PR#32, I#27). `shellcheck -x` was added in [C 07a3f34] (PR#36). `refuse-commit-to-main` was added in [C a184888] (PR#78). No job was ever removed or skipped by config. The hooks call tools directly, without `mise exec`, and assume mise is active in the shell [I#27 resolution].

Timings from lefthook's summary lines in the transcripts:

| Hook | n | median | p90 | max |
|---|---|---|---|---|
| pre-commit (whole hook) | 126 | 0.41 s | 1.22 s | 7.36 s |
| pre-push `test-unit` | 56 | 8.3 s | 31.5 s | 59.4 s |
| pre-push `lint-changed-packages` | 29 | 2.0 s | 5.0 s | 12.4 s |

### pre-commit: refuse commits on `main`

- **Enforced:** `scripts/refuse-commit-to-main.sh` fails when `HEAD` is `refs/heads/main`. A detached HEAD, as during a rebase, passes.
- **History:** at first only pushes to `main` were refused. The commit hook was added with no linked issue in [C a184888] (PR#78). The trigger is recorded in I#131: "In 4 sessions the agent committed on `main` and branched only after the pre-commit hook refused it" (out of the last 10).
- **Evidence:** `refuse-commit-to-main` failed in 16 sessions [T]. I#131 / [C c5b7910] then added an AGENTS.md line that names the behaviour ("branch before the first commit") and points to `docs/toolchain.md`, so agents don't only learn the rule from refusals.
- **Lesson:** this rule earned its keep. It's cheap (one `git symbolic-ref`) and catches a mistake agents really make. The hook catches it, and one line in AGENTS.md stops agents walking into it.

### pre-commit: `golangci-lint fmt {staged_files}` with `stage_fixed`

- **Enforced:** formats staged Go files and restages them.
- **History:** unchanged since [C 5dfd14b]. Before merge, partial staging was tested by hand: lefthook hides unstaged hunks, and when a formatter fix and an unstaged edit touch the same line it aborts with `conflict while merging unstaged changes` [I#27].
- **Evidence:** no `conflict while merging unstaged changes` appeared in any transcript [T]. Agents mostly committed with `git add -A` or whole files.
- **Lesson:** formatter-plus-restage worked for Go. *Inference:* the template needs the same for each language it formats (Go, `templ fmt`, the TS formatter from ADR-0013, the shell formatter if any). Each is a separate lefthook job with its own glob, and the stage-fixed behaviour only needs re-testing for formatters that rewrite more than whitespace.

### pre-commit: shellcheck on staged `*.sh`

- **Enforced:** `shellcheck -x {staged_files}`.
- **History:** `-x` was added in [C 07a3f34] because "staging any script that sources `lib.sh` failed on SC1091".
- **Evidence:** one pre-commit shellcheck failure in the transcripts [T]. CI also runs `make lint-shell` on every tracked script [C 8cd43ac].
- **Lesson:** keep it, with `-x` from the start if scripts share a sourced library.

### commit-msg: `cog verify --ignore-fixup-commits --file {1}`

- **Enforced:** Conventional Commits on each new commit message. Default types, no `cog.toml`. commitlint was rejected because it needs Node [I#27].
- **History:** unchanged.
- **Evidence:** the `conventional-commit` job failed 3 times in the transcripts [T]:
  - twice with `sh: cog: command not found` (exit 127) on 2026-10-01, in sessions that hadn't activated mise [T 6292ac63 00:05, a31c3f4a 00:30]. The agent got around it by wrapping commits in `zsh -lic '...'`.
  - once on a merge commit, `Merge branch 'main' into feat/107-edit-snippet` ("Missing commit type separator") [T agent-a6edd1f 2026-10-05T09:48]. In another session the agent wrote the merge message as `chore: merge main into feat/113-…` instead [T 33aff5c3 2026-10-06T02:17].
- Agents wrote Conventional Commit messages almost every time. The rule mattered more for PR titles, because the title becomes the squash commit (see `pr-title`).
- **Lesson:** cheap and rarely in the way. Two things to settle: hooks must find their tools without an activated interactive shell, and merge commits need a convention. Merging `main` into a branch was common, because force-push was denied (see the command guard).

### pre-push: refuse pushes to `main`

- **Enforced:** `scripts/refuse-push-to-main.sh` reads pre-push stdin and fails on `refs/heads/main`.
- **Evidence:** in the transcripts it fired only while it was being tested [T 6292ac63 2026-10-01T00:14]. The `main` ruleset already required PRs, squash merges and linear history, blocked force-pushes and deletion, and had **no bypass actors** [I#9]. The agent command guard denied pushes to `main` earlier still (see below).
- **Lesson:** a redundant third layer. It costs nothing, but the ruleset is what actually protects `main`.

### pre-push: `golangci-lint run` on changed packages

- **Enforced:** lints packages changed since the merge-base with `origin/main` (`scripts/changed-packages.sh`), and skips when no Go changed. Whole packages, because golangci-lint type-checks a package. `--new-from-rev` was rejected because it hides issues on old lines [I#27].
- **Evidence:** no failures of this job appear in the transcripts [T]. By push time the post-edit hook and the Stop gate had already linted the code.
- **Lesson:** cheap and quiet. Its main use is for code that never passed the agent hooks, which was the reason given for adding it: "Human-written code never passes the agent Stop gate" [C 5dfd14b].

### pre-push: `make test-unit` on every package

- **Enforced:** `go test -race ./...` (short) on every push, whatever changed.
- **Evidence of friction:**
  - Five pushes were blocked by `test-unit` failures. All five branches changed **no Go code**: research docs, the roadmap, a workflow diagnostic [T agent-a207de 2026-10-01T02:40, b5231aef 2026-10-06T02:37, agent-af5d84 2026-10-08T03:01, agent-abe41e 2026-10-08T03:06, e252affd 2026-10-09T00:05]. In the cases with a retry, the second push passed ("the retry passed from cache", [T b5231aef]). Several parallel agents were running test suites at the same time. I#154, I#160 and I#166 document load-dependent test flakes from the same days.
  - **The hook ran the hook tests against the real repo.** `scripts/agent/hooks_test.go` copied `os.Environ()` into its sandbox. Under pre-push, `GIT_DIR` was exported, so the sandbox's `git init`/`commit`/`switch` ran on the outer repo. It set `core.bare=true`, created a `feature` branch, and moved the pushed branch and local `origin/main` onto a commit that deleted every file. The push was rejected, and the state was repaired by hand [C c9a0e1c] (PR#49). The transcript shows the stray `chore: init` commit by `Test <test@example.com>` [T agent-a207de 2026-10-01T02:41].
- **Evidence of value:** none that only this job caught was found. Code-change pushes had already passed the Stop gate or the agent's own `make test`.
- **Lesson:** running all unit tests on every push gives a fixed delay (median 8 s, tail near 1 min) and lets flaky tests block pushes that can't have caused them. *Inference:* scoping to "Go or test inputs changed" (as the lint job already does), or leaving the full run to CI, would have avoided all five blocks. Separately, any test that shells out to git must strip `GIT_*` variables, because hooks export them.

### pre-push: feature and e2e only with `TUISNIP_SLOW_TESTS=1`

- **History:** the agent hooks design assumed pre-push would run them. Making them opt-in was a deliberate change, because "They can take minutes" [C 5dfd14b], [I#29 handoff].
- **Evidence:** the only agent push or commit with the variable set was in the session that built the hook [T 6292ac63 2026-10-01T00:14–00:19]. One other session set it on a plain `go test` run, not a push [T b5231aef 2026-10-06T04:34]. CI is where these tiers always ran [I#28].
- **Lesson:** an opt-in tier nobody opts into is documentation, not a guardrail. That was fine here, because CI required them.

## Agent hooks (`scripts/agent/`, `.claude/settings.json`)

Design: I#4 (research), I#13 (spec), built in [C f7b5ab5] (PR#35, I#29). Core bash scripts return 0 pass / 2 deny / 3 ask with the way forward on stderr. One thin adapter per agent system. Claude Code wiring: PreToolUse `Edit|Write|NotebookEdit` → path and comment guards, PreToolUse `Bash` → command guard, PostToolUse `Edit|Write` → `check-go-file.sh` (120 s timeout), Stop → `verify-build.sh` (300 s).

The cost of keeping the hooks working, from history:

| Commit | Fix |
|---|---|
| [C 07a3f34] | `LEFTHOOK*` env bypasses added to the deny list |
| [C c9a0e1c] PR#49 | hook test sandbox leaked into the real repo under pre-push |
| [C 326fbc9] PR#55 | comment guard (new), with a false positive on `" // "` in Go strings fixed during development |
| [C 339bd34] PR#59 | comment guard exempts `// ENUM(` |
| [C c7ec70d] PR#57 | copies, moves, links, `patch` and `git apply` into the repo denied |
| [C 67c3f07] PR#69 | in-place flags hidden in clusters (`perl -0pi`, `sed -nI`) |
| [C cdaa429] PR#70 | macOS awk crashed on non-ASCII text, which the adapter reported as a deny |
| [C a184888] PR#78 | `git commit -m"Add login"` denied as `commit -n` |
| [C a4de412] PR#90 | comment guard exempts `//name:arg` directives (`//sumtype:decl`) |
| [C 9658b40] PR#137 | per-edit lint added |
| [C 1c9a905] PR#169 | hooks judged the wrong checkout in linked worktrees |

`guard-command.sh` grew from 415 to 716 lines, and `hooks_test.go` from 613 to 1,063 lines, over these commits (line counts at [C f7b5ab5] and [C 1c9a905]).

### Path guard: deny generator output, ask on guardrail files

- **Enforced:** `protected-paths` denies sqlc output, `languages_gen.go`, mockery mocks, `*_enum.go`, `go.sum` and the generated `opencode.json`, plus any file carrying a `// Code generated … DO NOT EDIT.` header. It asks before edits to `scripts/agent/*`, agent wiring, `lefthook.yml`, `.golangci.yml`, `.go-arch-lint.yml`, workflows and `mise.toml` [I#13]. Each deny names the generator to run [I#29].
- **Evidence:** in 10 days of transcripts the deny tier **never fired**. Every PreToolUse Edit/Write deny was the comment guard [T]. The ask tier produced roughly 150 prompts, most of them for `scripts/agent/*` while the hooks themselves were being built (count approximate; the ask reason text also appears in file reads). I#29: "an ask from a hook prompts even in auto mode, once per matching edit. That only matters when an agent edits the guardrail files themselves, which was constant while building them. The user is keeping the ask tier as designed."
- **Lesson:** cheap to keep, but unproven as a guard. Agents may simply not try to edit generated files, or the drift check in CI (`make generate` + `git diff --exit-code`, [C 8cd43ac]) catches it anyway. The ask tier costs attention only during hook or CI work.

### Comment guard (`guard-go-comments.sh`)

- **Enforced:** the first submission of an edit that adds a Go comment is denied with the comment rule. The same comment submitted again in the session asks the user [I#53], [C 326fbc9].
- **History:** loosened twice for false positives: go-enum `// ENUM(` [C 339bd34] and `//name:arg` directives [C a4de412]. A third false positive (`" // "` inside a string literal) was fixed before merge [C 326fbc9].
- **Evidence:** about 46 first denials in 28 sessions. In about 32 of them the agent resubmitted the same comment line within its next 40 tool calls, which triggers the user prompt. In 14 it dropped or reworded the comment [T]. Whether the user approved each prompt isn't recorded.
- **Lesson:** it worked as a speed bump, and about a third of flagged comments went away. *Inference:* the rest became user prompts, so the rule's cost falls on the human. The user's global CLAUDE.md has the same "comment only why" rule. Directive-style comments exist in every generator the template uses (`//go:generate`, `//nolint`, sqlc and templ markers), so an exemption list would be needed from day one.

### Command guard (`guard-command.sh`)

Denies seen in the transcripts, by rule:

| Rule | Denies | Sessions | Real catch or false positive |
|---|---|---|---|
| interpreter writes a file | ~53 | ~41 | mostly real: agents writing via python/heredoc |
| redirect / `tee` outside `/tmp` | ~42 | ~33 | mostly real, some relative-path-after-`cd` cases ([C c5b7910] pitfalls list) |
| `sed -i` | ~40 | ~30 | real |
| `perl -i` | ~8 | ~8 | real |
| `mv`/`cp` into repo | ~6 | ~5 | real |
| awk `towc` crash on non-ASCII | 13 | 11 | false positive, fixed in [C cdaa429] |
| push to `main` | 8 | 4 | all false positives: agent in a worktree on a branch, hook `.cwd` on `main` (I#149) |
| force-push | 3 | 3 | 2 were `--force-with-lease` on the agent's own branch after a rebase |
| hook bypass (`--no-verify`, `LEFTHOOK=`, `core.hooksPath`) | 3 | 3 | all false positives (below) |
| `git reset --hard` | 1 | 1 | resetting a freshly created worktree |

- **Hook bypass:** no real bypass attempt was seen. The three denials were: `--no-verify` on a commit inside a throwaway `mktemp -d` repo used to test `refuse-commit-to-main.sh` [T 330f9dab 2026-10-02T11:05]; `printf 'LEFTHOOK=0 git commit -m x' | guard-command.sh`, a string piped to the guard to test it [T 90cdbbde 2026-10-01T02:08]; and `git config core.hooksPath` with no value, a read [T fb905285 2026-10-05T21:17]. [C 07a3f34] notes that quote flattening makes a quoted `LEFTHOOK=` argument a false positive. I#73 records `git commit -m"Add login"` denied as `commit -n`.
- **Push to `main`:** reworked in [C 1c9a905] (PR#169, I#149) so it follows `cd` and `git -C` and judges the checkout the command runs in. I#149: "Four of the agents that worked on #116, #117, #118 and #119 in parallel worktrees reported" the post-edit failure, and the push denial recurred with worktrees outside `.claude/worktrees/`.
- **Shell file writes:** this rule exists so file changes go through the edit tools, where PostToolUse format and vet run [I#13]. I#131: "In the last 10 sessions, the command guard blocked shell file writes about 25 times." The fix was documentation (an AGENTS.md workflow line and a "Shell pitfalls" list), not loosening the rule [C c5b7910].
- **Parser cost:** every guard bug fixed was in shared lexing (quote flattening, flag clusters, locale), not in a policy rule: "Both recent fixes (#69, #70) went into the shared lexing and parsing code, not into a policy rule. Gremlins only mutates Go, so the shell guard isn't covered by mutation testing" [I#73]. I#73 proposed moving the lexer to Go (`mvdan.cc/sh`). It was never done [C a184888 "Not in this PR"]. The guard documents itself as "not a security boundary": `sh -c '…'` isn't inspected [docs/toolchain.md].
- **Lesson:** the shell-write deny did its job: it kept edits flowing through the post-edit checks. But a bash command parser keeps attracting edge-case bugs, and each false positive costs an agent retry. The git-safety rules mostly produced false positives, because git hooks and the server ruleset already covered the same ground. The force-push deny and the commit-msg rule together pushed agents towards merge commits.

### Post-edit check (`check-go-file.sh`)

- **Enforced:** `golangci-lint fmt <file>`, then `go vet -tags feature,e2e ./<pkg>`, and since [C 9658b40] `golangci-lint run --allow-serial-runners ./<pkg>`.
- **History:** lint was added because "golangci-lint ran only in the end-of-turn Stop gate and in pre-push … In the last 10 sessions that first run reported between 3 and 35 issues … Several of those forced redesigns: one deleted 4 files and rewrote them, another changed 20 call sites" [I#129]. Measured cost: 7.6 s cold, about 1.2 s warm for the largest package [C 9658b40]. `--allow-serial-runners` was added because parallel runs failed on golangci-lint's lock. The worktree fix came in [C 1c9a905].
- **Evidence:** the busiest hook by far [T]. About 1,360 blocking vet reports in about 40 sessions (about 507 after editing a `_test.go` file, about 562 citing `undefined:`), and 11 `no non-test Go files` false positives on test-only packages. Since 2026-10-05, 108 lint reports in 10 sessions, led by `unused` (32), `revive` (22), `exhaustruct` (15), `wsl` (14), `unparam` (14) and `iface` (13). Every lint report started with three `Skipped 0 issues by rules` warnings, noise first logged in I#27 and never removed. The formatter's rewrites produced about 430 "file modified by hook" notices, which Claude Code absorbs without a failed edit.
- *Inference:* most vet failures were intermediate states in multi-edit changes and TDD red phases. The design said "Tests don't run here, so TDD's red phase isn't reported" [I#13], but vet type-checks test files, so a red test that references a not-yet-written symbol does get reported. Similarly, `unused` and `unparam` at edit time flag code whose caller hasn't been written yet.
- **Lesson:** per-edit feedback is where most hook value came from. I#129 is the clearest evidence in the repo that late lint findings cost rework. It's noisy during multi-step edits, and it must resolve the module from the edited file's path, not the hook's cwd.

### Stop gate (`verify-build.sh`)

- **Enforced:** on changed packages (committed and uncommitted since the merge-base): `go build ./...`, `go-arch-lint check`, `golangci-lint run`, `go test -short -race` (no `./test/...`). It skips when the Go diff fingerprint matches the last pass, and releases after 3 consecutive blocks [I#29].
- **Evidence:** only 12 blocks in 5 sessions: 1 build failure, the rest lint [T]. It was released after 3 blocks 4 times. Three of those were in one session, where the failing file belonged to **another session working in the same checkout**. The agent said: "the Stop hook lints the whole working tree, not just the files this session changed. So two agents working in the same checkout keep tripping each other's hook" [T acfcf6ca 2026-10-06T11:56]. One Stop run emitted `fatal: this operation must be run in a work tree` [T 95e66892 2026-10-01T02:43]. *Inference:* this is the same morning the hook test sandbox set `core.bare=true` on the repo [C c9a0e1c]. `verify-build.sh` was left unchanged by the worktree fix: it "still checks `.cwd`'s checkout" [C 1c9a905].
- **Lesson:** once the post-edit lint existed, the Stop gate rarely had anything left to catch. Its scope is "everything uncommitted in this checkout", not "this session's edits", which breaks when sessions share a checkout. The 3-block release valve did its job.

### Multi-agent adapters (OpenCode, Antigravity)

- **Evidence:** "The OpenCode plugin was exercised under Bun against the real scripts, not inside OpenCode. Antigravity is untested live" [I#29]. Comment-guard support there rests on an "unverified" field name [I#53]. No transcripts exist for either system.
- **Lesson:** every core-script change was carried through three adapters and their fixtures without any use of the other two. This is maintenance cost with no observed benefit.

## CI and the `main` ruleset

- **`ci.yml`** [C 8cd43ac]: lint (+ `fmt --diff`), arch-lint, test (unit, feature, e2e), a 4-target build matrix, govulncheck, a `generated` drift check, shellcheck. `test-macos` was added with the clipboard code [C 4e14da0]. All of them are required checks on the ruleset, with "Require branches to be up to date" off [I#28]. Every job calls a Make target so it can run locally [C 8cd43ac]. Flakes did reach CI: `test-macos` failures in PR#127 and I#160. Their causes were in the tests (load-dependent timing, an upstream Bubble Tea race) [I#160], [C d0826d0].
- **`pr-title.yml`**: `cog verify "$PR_TITLE"`, with the title passed through an env var. It is its own workflow, so a title edit re-runs only it. It catches what the commit-msg hook can't, because with squash merges the PR title becomes the commit on `main` [I#28]. It failed PR titles without a Conventional Commit prefix in 2 of 10 sessions [I#131]. `cog` needs a git identity even to verify a string [I#28].
- **Deep tests:** nightly became weekly the same day: "Daily runs would mostly search unchanged code" [C c80dfab], [I#28 follow-up]. Mutation testing was first required before every PR. Agents skipped it or ran it wrongly [C 8e78671]. Then the pre-PR requirement was dropped because runs froze the dev machine, and the CI job moved to manual-only after it ran the runner out of memory [I#161], [C a01c76f].
- **Ruleset:** PR required, squash only, linear history, no force-push or deletion, no bypass actors [I#9]. PR#169 relies on it: "A wrong pass is still caught by the git hooks and CI."
- **Lesson:** CI plus the ruleset gave the hard guarantees. Local hooks and agent hooks only moved feedback earlier. Scheduled heavy jobs were cut back as soon as their reports went unread.

## Cross-cutting findings

1. **Worktrees broke the agent hooks.** Every hook that worked from the hook's `.cwd` misjudged agents in linked worktrees: wrong-module vet, false push-to-`main` denials, wrong-checkout Stop gate [I#149], [C 1c9a905]. Parallel agents in worktrees were the normal way of working by 2026-10-05.
2. **Two layers of git safety were mostly redundant with the ruleset.** Push-to-`main` and hook-bypass rules produced only false positives in the agent hook, and only self-tests in the pre-push hook. The ruleset had no bypass actors.
3. **Moving feedback earlier paid off most.** Per-edit lint [I#129], the shell-write deny that keeps edits on the checked path [I#13], and AGENTS.md lines that state the behaviour up front [I#131] were the changes made in response to friction, and none were reverted.
4. **Tests that run in hooks inherit hook state.** `GIT_DIR` from pre-push [C c9a0e1c], and BSD awk under a UTF-8 locale on macOS that CI's mawk doesn't reproduce [C cdaa429].
5. **The unit-test pre-push job was the only hook that blocked unrelated work.** Five docs-only pushes were blocked by flaky or leaking tests [T].
6. **Hook tools must resolve without shell activation.** `cog: command not found` on agent commits [T], from the "hooks call tools directly … assume mise is activated" choice [I#27].

## What each lesson means for this template

The template is Go + templ + TypeScript (Vite+) + SQL (goose, sqlc) + shell, with Make as the single command surface [ADR-0010]. Tools are pinned in `tools/tools.mod`, separate modfiles or `bin/` rather than mise [ADR-0007], [ADR-0010]. Tests are split by build tag, and integration and feature tests need Docker [ADR-0011]. The notes below are implications for the hook-rules ticket, not decisions.

- **pre-commit formatting.** TuiSnip had one formatter. The template has at least Go, `templ fmt`, the Vite+ formatter (oxfmt via `bin/vp`, [ADR-0013]) and possibly shell and SQL formatters. ADR-0010 already says lefthook is the only hook runner and `vp staged` isn't used. *Inference:* each formatter becomes a lefthook job with its own glob and `stage_fixed`. TuiSnip's partial-staging test only covered `golangci-lint fmt`.
- **Hooks calling Make targets.** ADR-0010 says "pre-push, CI and the Dockerfile call those same targets". TuiSnip's `cog: command not found` shows why: a hook that calls `make <target>`, with tools under `bin/` or `go tool -modfile`, doesn't depend on the agent's shell having mise or anything else on PATH. No template ADR pins `cog` yet.
- **`cog` commit-msg.** Low friction in TuiSnip. Needs a merge-commit convention if agents merge `main` into branches, which they did because force-push was denied.
- **Refusing commits/pushes on `main`.** The commit refusal caught real mistakes. The push refusal was a backstop to a ruleset. Whether the template's GitHub repo gets the same no-bypass ruleset decides whether the push refusal matters.
- **pre-push split.** TuiSnip's unit tests on every push blocked docs-only pushes. ADR-0011 makes unit and architecture tests Docker-free, and JS unit tests join `make test`. Integration and feature tests need Docker, so they fit the same opt-in-or-CI-only slot as TuiSnip's `TUISNIP_SLOW_TESTS`. *Inference:* a changed-files skip, like the lint job's, applied to Go, `.templ`, `.sql` and TS inputs, would have avoided every observed block.
- **Generated files.** The template has more generator output than TuiSnip (`*_templ.go`, sqlc output, mockgen mocks, `migrations.sum`, `internal/asset/dist/`). TuiSnip's deny tier never fired, and its CI drift check covered the same ground [C 8cd43ac]. Whether `*_templ.go` carries the standard `DO NOT EDIT` header that TuiSnip's header check relies on wasn't checked here.
- **Post-edit checks per language.** TuiSnip's highest-value agent hook was per-edit fmt, vet and lint on the edited Go package. For `.templ` edits, `go vet` needs `templ generate` first, or it vets stale `_templ.go`. For TS, the equivalent is the Vite+ lint and type-check targets. Neither was tried in TuiSnip. The check must resolve the module from the edited path [C 1c9a905], and must expect red-phase vet failures on test files.
- **Command guard.** The shell-write rule did its job, at the cost of a 716-line bash parser with recurring false positives [I#73]. Its git-safety rules duplicated git hooks and the ruleset. A smaller rule set, or a Go lexer as I#73 proposed, are the options the evidence points to. This repo's own research flow (this ticket) runs agents in `git worktree add` checkouts, the case that broke TuiSnip's guards.
- **Comment guard.** The rule matches the user's global CLAUDE.md. About two thirds of flagged comments were resubmitted to the user. templ, sqlc and `//go:generate` all use directive-style comments that would need exemptions.
- **Stop gate.** Rarely fired once per-edit lint existed, and misfired when sessions shared a checkout. With Docker-dependent test kinds [ADR-0011], it couldn't run integration or feature tests anyway.
- **Agent systems.** TuiSnip's OpenCode and Antigravity adapters were never used live. The ticket names only Claude Code hooks.
- **Tests of hook scripts.** If the template's hook scripts get Go tests that shell out to git, they must drop `GIT_*` variables [C c9a0e1c]. Shell tested only on Linux CI misses BSD tool behaviour on macOS [C cdaa429].

## Not established

- Whether the user approved or rejected each comment-guard and guardrail-file ask prompt. The transcripts record the prompt, not the answer.
- Why the five pre-push `test-unit` runs failed (other than the 2026-10-01 sandbox leak). The failing test names weren't captured in the output the agents kept.
- Whether the Stop hook ran for subagents. Only 2 of 131 subagent transcripts contain Stop-hook feedback, and the reason wasn't checked.
- Whether any human (non-agent) push or commit used `--no-verify`. Git doesn't record it.
