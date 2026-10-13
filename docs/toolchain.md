# Toolchain

Every tool the repo uses: its version, its role and the pitfalls to know about. The ADRs explain why each tool was chosen. [upgrading.md](upgrading.md) explains how to bump the pinned ones.

## Go

| | |
| --- | --- |
| Version | `go 1.27`, `toolchain go1.27.2` in `go.mod` |
| Role | Compiles the app. Each process is its own `main` under `cmd/` ([ADR-0010](adr/0010-makefile-as-the-command-surface-with-one-binary-per-process.md)). |

- Any Go 1.21 or newer reads the `toolchain` line and downloads that exact version, so the Go you have installed only needs to be recent enough to bootstrap.
- `GOTOOLCHAIN=local` turns that download off, and the build then fails if your local Go is older than the `go` line.

## Make

| | |
| --- | --- |
| Version | GNU Make 3.81 or newer (macOS ships 3.81) |
| Role | The only command surface. Run `make` to list the targets ([ADR-0010](adr/0010-makefile-as-the-command-surface-with-one-binary-per-process.md)). |

- `make` with no target runs `help`, which prints every target that has a `## description`, grouped under the `##@ Group` headers above it. The parsing lives in `scripts/make-help.awk`. Targets without a `##` comment are internal and stay hidden.
- `.env` is included if it exists but not exported, so recipes see the same overrides as Compose and the app.
- Composite targets only call other targets, and each leaf target runs one command.
- The Go targets work in the module `GO_MODULE` names, relative to the checkout (`.` by default), and read `FILES` relative to it. `tools/agenthook` is a module of its own, so `lint-agenthook` and `test-agenthook` run `lint` and `test-go` with `GO_MODULE=tools/agenthook`.
- Make 3.81 has no `.ONESHELL` and no `--output-sync`.
- Make 3.81 reads an escaped colon in a prerequisite (`a\:b`) literally, backslash included, so it never matches the target. Target names use hyphens instead of colons.

## Pinned binaries

| | |
| --- | --- |
| Version | `tools/<tool>.version`, with `tools/<tool>.sha256` in `sha256sum` format |
| Role | Tools that aren't Go modules are downloaded into `bin/` by a real Make file target, defined once as the `pinned-binary` macro and implemented by `scripts/install-pinned-binary.sh`. |

- The target detects OS and architecture, downloads with `curl`, refuses to install an asset whose checksum doesn't match, and re-fetches when the version file, the checksum file or the script changes.
- Each binary gets a `bin/<tool>.pin` with the asset URL and its pinned checksum. When Make asks for a binary whose pin still matches, the script only refreshes its timestamp. That is what makes CI's restored `bin/` cache useful, because a checkout leaves the version and checksum files newer than every cached binary.
- Only macOS and Linux on arm64 and x86_64 are supported. Hand-built checksum files list only those, while a verbatim upstream copy such as `tools/vp.sha256` lists every asset. Other platforms fail with "no checksum" or a missing asset.
- Make 3.81 compares timestamps to the second, so a version file edited in the same second as the last download doesn't trigger a re-fetch.

## golangci-lint

| | |
| --- | --- |
| Version | v2.14.0, alone in `tools/golangci-lint.mod` |
| Role | Lints (`make lint`) and formats (`make fmt`, `make fmt-check-go`) Go code. Configured in `.golangci.yml` ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)). |

- It runs through `go tool -modfile=tools/golangci-lint.mod`, so the project toolchain compiles it and it never fails with "built with an older Go". The first run compiles it and is slow.
- `linters.default: all` turns on every linter, including deprecated ones that have a successor. Each disabled linter has its reason next to it.
- By default only one issue is shown per line. `--uniq-by-line=false` shows all of them.
- `run.allow-serial-runners` makes a run wait for another one's lock instead of failing, because the agent hooks lint while other sessions or a human lint too.
- The formatters are gofumpt with its extra rules, goimports with the module path as the local prefix, and golines at 120 columns. `lll` is off because golines wraps long lines.
- The only escape hatch is `//nolint:<linter> // <why>`. nolintlint rejects a directive without a specific linter or a reason.

## go-arch-lint

| | |
| --- | --- |
| Version | v1.19.0, alone in `tools/go-arch-lint.mod` |
| Role | Checks imports between components and confines each third-party library to the components that may use it (`make arch`). Configured in `.go-arch-lint.yml` ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)). `make arch-graph` draws the components into `tmp/arch-graph.svg`. |

- It is a whitelist: a Go file outside every component fails the check, and so does a vendor import that no `canUse` allows. The standard library is always allowed, so stdlib bans live in depguard.
- Any warning fails the check. `--max-warnings` only limits how many are printed, and it rejects 0.
- A package that matches several components belongs to the most specific one, so `internal/adapter/postgres` is `adapter-postgres`, not `adapter`.
- A component may not import its own subpackages unless it lists itself in `mayDependOn`.
- `ignoreNotFoundComponents` is on, so components can be declared before their packages exist.
- It reads `_test.go` files like any other file, so test-only libraries are allowed everywhere here and kept out of non-test code by depguard. `testdata/` and `node_modules/` directories are excluded, because some npm packages ship Go files.
- deepScan is off ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)). Turning it back on needs a `golang.org/x/tools` newer than the v0.49 go-arch-lint v1.19.0 requires: that version cannot read Go 1.27 export data, and deepScan then silently checks nothing.

## Architecture tests

| | |
| --- | --- |
| Version | stdlib `go/parser` and `go/ast` |
| Role | Check code shape in `tests/architecture/`: the Action shape, no exported interfaces in `internal/`, no functional options, no `go` statements in the core layers. They run with `make test-go` and `make test`, and need no Docker. |

- Test files and generated files are skipped.
- An exception is a named entry in the allowlist inside the test, keyed by the subject the violation prints. An entry that matches no violation fails the test.
- Each rule runs against a fixture in `tests/architecture/testdata/<rule>/` that proves it catches its violation.

## shellcheck

| | |
| --- | --- |
| Version | `tools/shellcheck.version`, a [pinned binary](#pinned-binaries) in `bin/shellcheck` |
| Role | Lints every tracked `*.sh` file (`make lint-shell`), following `source`d files (`-x`). pre-commit lints only the staged scripts. |

- Untracked scripts and scripts without the `.sh` extension are not linted.

## lefthook

| | |
| --- | --- |
| Version | v2.2.1, in `tools/tools.mod` ([ADR-0010](adr/0010-makefile-as-the-command-surface-with-one-binary-per-process.md)) |
| Role | Runs the git hooks configured in `lefthook.yml`. `make setup` installs them into `.git/hooks`. |

- Every job calls a Make target, so the hooks need nothing on `PATH` beyond the system tools and Go.
- The `lefthook` key in `lefthook.yml` makes the installed hooks run the pinned version through `go tool`. Without it they would use whichever lefthook is on `PATH`.
- `glob_matcher: doublestar` makes `**/*.go` match files at the root too. The default matcher needs at least one directory.
- pre-push jobs list their files with `scripts/changed-files.sh` (changed since the merge-base with `origin/main`) and filter them with their own `glob`. A job is skipped when no file matches. A new job only declares its globs.
- Without `origin/main`, as on the first push to an empty remote, the helper lists every tracked file, so every job runs.
- `LEFTHOOK=0 git commit …` skips the hooks. `LEFTHOOK_VERBOSE=1` prints each command a job runs.
- Hooks export `GIT_DIR` and friends. A test that shells out to git strips `GIT_*` from its environment and works in `t.TempDir()`, as `tests/hooks/` does.

## cog

| | |
| --- | --- |
| Version | `tools/cog.version`, a [pinned binary](#pinned-binaries) in `bin/cog` |
| Role | Checks that commit messages (`make lint-commit-msg COMMIT_MSG_FILE=<path>`) and pull request titles (`make lint-pr-title PR_TITLE=<title>`) are Conventional Commits. |

- It uses the default commit types. There is no `cog.toml`.
- In commit messages, `fixup!` and merge commits pass, and `#` comment lines in the message file are ignored. A pull request title passes neither way, because the squash merge makes it the commit on `main`.
- Linux x86_64 gets the musl build and Linux aarch64 the gnu build, because those are the only Linux builds upstream ships for each.

## gitleaks

| | |
| --- | --- |
| Version | v8.30.1, alone in `tools/gitleaks.mod` |
| Role | Scans the staged changes for secrets (`make lint-secrets`). Configured in `.gitleaks.toml`. |

- `.gitleaks.toml` extends the default rules and allowlists test fixtures by path (`*_test.go`, `testdata/`), so a TOTP seed in a test passes.
- Findings are printed with the secret redacted, as rule, file and line.
- A false positive outside the allowlisted paths takes a `gitleaks:allow` comment on its line.

## govulncheck

| | |
| --- | --- |
| Version | v1.9.0, alone in `tools/govulncheck.mod` |
| Role | Reports known vulnerabilities that the code in `GO_MODULE` can reach, in its dependencies and the standard library (`make vuln`). CI runs it on both modules. No hook runs it. |

- It only reports a vulnerability when the code calls the affected function. An imported but unreached one is listed as informational and does not fail the check.
- It reads the vulnerability database at vuln.go.dev, so it needs the network. A new advisory can fail CI on a branch whose code did not change.

## Vite+

| | |
| --- | --- |
| Version | `tools/vp.version`, a [pinned binary](#pinned-binaries) in `bin/vp`, and the `vite-plus` catalog entry in `pnpm-workspace.yaml` |
| Role | Builds the frontend assets (`make vite-build`, `make watch-vite`), formats with oxfmt (`make fmt-js`, `make fmt-check-js`), lints with oxlint (`make lint-js`), type-checks with `tsc` (`make typecheck-js`) and runs the JS unit tests with Vitest (`make test-js`). Configured in `vite.config.ts` ([ADR-0013](adr/0013-vite-plus-frontend-toolchain-with-a-build-watch-dev-loop.md)). |

- `tools/vp.sha256` is a verbatim copy of the release's `vp-checksums.txt`, so it also lists the Windows assets.
- Make runs `bin/vp` with `VP_HOME=.vite-plus`, which keeps Node, pnpm and vp's own shims inside the checkout. Without it, `bin/vp` sets itself up in `~/.vite-plus` and repoints a global vp's shims at itself. Running `bin/vp` by hand needs the same variable.
- The first run after `bin/vp` is fetched prints shell setup steps. Make never uses the shims, so they can be ignored.
- The Vite root is the repo root, so Tailwind scans every file git doesn't ignore, `.templ` included.
- `internal/asset/src/app.ts` is the only entry. It imports `app.css`, so the manifest lists the stylesheet under the entry's `css` key.
- The build empties `internal/asset/dist/` and writes the manifest to `dist/manifest.json`. A small plugin in `vite.config.ts` writes `.gitkeep` back, so the tracked file never shows as deleted.
- The modulepreload polyfill is off and `assetsInlineLimit` is 0, so the output holds no inline script and no `data:` URI.
- oxlint fails on the `correctness`, `suspicious` and `perf` categories, and on functions over the length, complexity, depth and parameter limits golangci-lint uses for Go.
- oxfmt skips Markdown and TOML, and sorts the keys of `package.json`.
- Tests import the Vitest API from `vite-plus/test`, not from `vitest`. They run under happy-dom and sit next to the code as `*.test.ts`.
- `typecheck-js` runs `tsc` from the `typescript` devDependency (TypeScript 7), not oxlint's type-aware mode.

## Node

| | |
| --- | --- |
| Version | `.node-version` |
| Role | Runs the Vite+ toolchain. `bin/vp` downloads that exact release into `.vite-plus/` on first use. |

- Vite+ 1.x needs Node `^22.18.0 || ^24.11.0 || >=26.0.0`.
- A Node on `PATH` is never used by the Make targets.

## pnpm

| | |
| --- | --- |
| Version | The `packageManager` field in `package.json`. `bin/vp` downloads it. |
| Role | Installs the JS dependencies into `node_modules/` (`make setup`). Configured in `pnpm-workspace.yaml`. |

- Every install uses `--frozen-lockfile`, so a `package.json` change without the matching `pnpm-lock.yaml` fails. Add or remove a dependency with `VP_HOME=.vite-plus bin/vp add -D <package>` or `VP_HOME=.vite-plus bin/vp remove <package>`, which update both.
- `minimumReleaseAge` holds back versions published less than three days ago. `vp migrate` exempts the Vite+ toolchain packages, so a new `bin/vp` can install its own.
- Dependency lifecycle scripts are blocked.

## Production image

| | |
| --- | --- |
| Version | `golang:<toolchain>-trixie` for the build stage, `gcr.io/distroless/static-debian12:nonroot` pinned by digest for the runtime, both in `Dockerfile` ([upgrading.md](upgrading.md#go-build-image)) |
| Role | `make image` builds the image for the host platform and tags it `IMAGE_NAME:VERSION` and `IMAGE_NAME:latest`. Nothing is pushed. |

- The build stage runs `make build`, which fetches `bin/vp` and Node, installs from the frozen lockfile, runs the Vite build and then compiles `cmd/web` with `CGO_ENABLED=0`, `-trimpath` and `-ldflags="-s -w -X …/buildinfo.Version=$(VERSION)"`.
- `VERSION` defaults to `git describe --always --dirty`. `.git` is not in the build context, so `make image` passes it in as a build arg and `make build` stamps it into `buildinfo.Version` with `-X`. A plain `docker build` without `--build-arg VERSION` fails.
- Cache mounts keep the Go module and build caches, the pnpm store and `.vite-plus/` between local builds. CI starts without them.
- `.dockerignore` lists what `.gitignore` does, plus `.git`. Change both together.
- The runtime runs as uid 65532 with CA certificates and tzdata, and holds `/app/web` with `ENTRYPOINT []` and `CMD ["/app/web"]`, so the platform can override the command without resetting an entrypoint.
- There is no `HEALTHCHECK`: distroless has no shell or curl. Orchestrators probe the app's HTTP endpoints instead.
- `IMAGE_NAME` is the template's slug, not `APP_NAME`, which `.env` would leak into the tag.

## Hooks

| Hook | Runs | Tool |
| --- | --- | --- |
| pre-commit | Refuses commits on `main`. A detached HEAD, as during a rebase, passes. | `make refuse-commit-on-main` |
| pre-commit | Formats the staged Go files and re-stages them | `make fmt` (golangci-lint) |
| pre-commit | Formats the staged TypeScript, CSS and JSON files and re-stages them | `make fmt-js` (oxfmt) |
| pre-commit | Lints the staged shell scripts | `make lint-shell` (shellcheck) |
| pre-commit | Scans the staged changes for secrets | `make lint-secrets` (gitleaks) |
| commit-msg | Checks the message is a Conventional Commit | `make lint-commit-msg` (cog) |
| pre-push | Refuses pushes to `main`, because GitHub doesn't enforce rulesets on private repos of free accounts | `make refuse-push-to-main` |
| pre-push | Lints the Go packages changed since the merge-base with `origin/main` | `make lint` (golangci-lint) |
| pre-push | Checks the architecture when Go, `go.mod` or `.go-arch-lint.yml` changed | `make arch` (go-arch-lint) |
| pre-push | Runs the unit, architecture and tooling tests when a `.go`, `.templ`, `.sql`, `.sh`, `go.mod` or `go.sum` file changed. Tooling tests cover the scripts. | `make test-go` |
| pre-push | Lints and tests the [agent hooks](#agent-hooks) when a file under `tools/agenthook/` changed. The three jobs above skip that module. | `make lint-agenthook`, `make test-agenthook` |
| pre-push | Type-checks, lints and runs the JS unit tests when a file under `internal/asset/src/`, `package.json`, the lockfile, `pnpm-workspace.yaml`, `.node-version`, `tsconfig.json` or `vite.config.ts` changed | `make typecheck-js`, `make lint-js`, `make test-js` |

pre-commit and pre-push are piped: the first failing job stops the rest. No hook runs integration, feature or e2e tests. CI runs those.

## Agent hooks

| | |
| --- | --- |
| Version | `tools/agenthook`, a Go module of its own, which pins `mvdan.cc/sh/v3` in its `go.mod` |
| Role | Guards and checks a Claude Code agent's edits and shell commands. `.claude/settings.json` calls `make -s agenthook-<event>`, which rebuilds `bin/agenthook` whenever its sources change and pipes the hook JSON through stdin. |

| Event | Rule | Change it in |
| --- | --- | --- |
| Before an edit | Denies edits to generated output and names what regenerates it: a file with a `Code generated … DO NOT EDIT.` header (`make generate`), `internal/asset/dist/` (`make vite-build`), `migrations.sum` (`make generate`), `go.sum` (`go mod tidy`) and `pnpm-lock.yaml` (`bin/vp add` or `remove`) | `internal/editrule/generated.go` |
| Before an edit | Asks before editing a guardrail file: `.golangci.yml`, `.go-arch-lint.yml`, `lefthook.yml`, `tools/agenthook/`, `.claude/settings.json`, `.github/workflows/`, `.github/rulesets/` | `internal/editrule/guardrail.go` |
| Before an edit | Asks before editing `go.mod`, `tools/*.mod` or `package.json`, which can add a dependency | `internal/editrule/guardrail.go` |
| Before an edit | Comment guard: a comment added to a Go, TS, templ, SQL or shell file is denied with "comments only explain a why the code cannot", and the same comment resubmitted in the session asks the human. Comments already in the file pass, and so do directives (see below). | `internal/comment/languages.go` |
| Before a shell command | Denies shell writes into a checkout, so every change goes through the edit tools and their post-edit checks: redirects, `tee`, `sed -i`, `perl -i`, `cp`, `mv` and `ln` into it, `patch`, `git apply`, and inline interpreter scripts (`-c`, `-e`, heredocs) that call a file-writing API. Writes outside every checkout, as under `$TMPDIR`, pass. | `internal/commandrule/write.go` |
| Before a shell command | Denies `git push` to `main`: a `main`, `HEAD:main` or `refs/heads/main` refspec, `--all`, `--mirror`, or a push without a refspec where `main` is checked out. Other branches pass, `--force-with-lease` included. | `internal/commandrule/push.go` |
| Before a shell command | Asks before `git reset --hard`, `clean -f…`, `checkout -- <path>`, `checkout .`, `restore` without `--staged`, `stash drop`, `stash clear`, `branch -D` and `worktree remove --force` | `internal/commandrule/destructive.go` |
| Before a shell command | Asks before `go get`, `go mod edit -require` or `-tool`, `pnpm add`, `vp add` and `npm install <package>`. `go mod tidy` and frozen installs pass. | `internal/commandrule/dependency.go` |
| After an edit | A `.go` file gets `make fmt`, `make vet` (every test tag) and `make lint-wip` (no `unused` or `unparam`, since the caller is often not written yet) in its module. A `.ts` file gets `make fmt-js` and `make lint-js`. The first failure goes back to the agent. | `internal/check/after_edit.go` |
| End of turn | `make build-packages`, go-arch-lint, `make lint` and `make test-go` on the packages this session edited, per module, then `make typecheck-js` and `make test-js` if it touched TypeScript. A failure blocks the turn from ending. The gate is skipped when nothing changed since it last passed, and it lets the turn end after 3 blocks in a row. | `internal/check/end_of_turn.go`, `internal/hook/gate.go` |

- Exempt directives: Go `//go:…`, `//nolint:<linter> // <reason>`, `//name:arg` directives and the generated-code header; TS `// @ts-expect-error`, `/// <reference …>` and `oxlint-disable`, `-line`, `-next-line` or `oxlint-enable` naming at least one rule; SQL `-- name: … :<cmd>` and `-- +goose …`; the shebang and `# shellcheck …`; and a `gitleaks:allow` marker in any language. A bare `oxlint-disable` and every `eslint-` directive count as comments, and oxlint's `unicorn/no-abusive-eslint-disable` holds humans to the same rule.
- The adapter in `internal/claude` is the only one, and the rules in `internal/hook` know nothing about Claude Code. Every decision goes out as Claude Code's JSON with exit status 0, so a non-zero status only means the hook broke. `settings.json` turns any failure into exit 1, which Claude Code reports without blocking, so a `bin/agenthook` that no longer builds cannot block the edit that fixes it.
- Every rule judges the checkout of the edited path, or of the directory a command runs in after its `cd`s and `git -C`s, never the hook's working directory. A linked worktree gets the checks of its own Makefile and module.
- Session state (edited paths, comments already denied, the gate's counters) lives in `$(git rev-parse --git-dir)/agenthook/<session id>.json`, so linked worktrees and parallel sessions never share it. The end-of-turn gate reads this session's state from every worktree of the repository.
- go-arch-lint checks the whole root module and cannot be limited to packages, so the gate runs `make arch-json` and keeps only the warnings on this session's files. Two sessions editing different Go packages in one checkout never fail each other's gate. `typecheck-js` and `test-js` cover the whole frontend, so a session that touched TypeScript also sees another session's TypeScript failures.
- The gate records the edits it let through, after a pass or a release, and stays quiet until they change.
- The comment guard lexes rather than parses, so it works on fragments. A line comment marker inside a word, as in `https://` or `${#x}`, does not start a comment, and templ treats only `"` and `` ` `` as string quotes, because its HTML text is full of apostrophes.
- Shell commands are parsed with `mvdan.cc/sh`, never run. Words expand with the hook's environment plus the variables the script assigns. Of the command substitutions, only `$(pwd)` and `$(git rev-parse --show-toplevel)` resolve. Any other substitution stays unknown, and so does a path or directory built from it, which the rules then pass.
- Interpreter scripts are matched against known file-writing calls, not run. A write the list does not know passes, and a script file such as `python3 x.py` is not read.
- No rule covers hook bypasses or force-pushes: the git hooks, the ruleset and CI cover them.
- The tests are tooling tests: table tests over the shell parser and every rule decision in `internal/`, plus recorded Claude Code hook JSON, kept as `testdata/*.json.tmpl` templates that fill in the paths, replayed through the built binary against throwaway repos whose stub Makefile logs the targets it is asked to run.

## CI

| Job | Runs |
| --- | --- |
| `lint` | `make fmt-check`, `lint`, `lint-agenthook`, `lint-js`, `typecheck-js`, `lint-shell` and `arch` |
| `generate` | `make generate`, then `make drift-check`, which fails if that changed a tracked file or created a new one |
| `test` | `make test` |
| `vuln` | `make vuln` and `make vuln-agenthook` |
| `image` | `make image`, which builds the [production image](#production-image) without pushing it |
| `pr-title` | `make lint-pr-title`, which runs `cog verify` on the pull request's title |

- `.github/workflows/ci.yaml` runs the first five jobs in parallel on pull requests and on pushes to `main`. A newer run cancels an older one on the same ref.
- `pr-title` has its own workflow, `.github/workflows/pr-title.yml`, which also runs when a pull request is edited. Editing a title re-runs only that check.
- Every job runs on `ubuntu-24.04` with `permissions: contents: read`, a checkout without persisted credentials and a timeout: 10 minutes, 20 for `test` and 5 for `pr-title`.
- Every job calls only Make targets ([ADR-0010](adr/0010-makefile-as-the-command-surface-with-one-binary-per-process.md)), apart from the `pr-title` step that gives `cog` a git identity. Steps that several jobs share are written once as YAML anchors.
- Actions are pinned by full commit SHA, with the version in a trailing comment. Nothing bumps them automatically ([upgrading.md](upgrading.md#github-actions)).
- Go comes from `actions/setup-go`, which reads the `toolchain` line of `go.mod`. Node comes only from `bin/vp` through `make js-install`, so `.node-version` stays its only pin. The `image` job sets up neither, because the Dockerfile's build stage brings both.
- Caches: the Go module and build cache, one per job and keyed on `go.mod`, `go.sum`, `tools/*.sum` and `tools/agenthook/go.sum`, because the jobs build different packages and tools while setup-go's own cache would share one key between them; `bin/`, keyed on `tools/*.version`, the checksum files and `scripts/install-pinned-binary.sh`; the pnpm store, keyed on `pnpm-lock.yaml`; and the golangci-lint cache, keyed on `tools/golangci-lint.sum` plus the commit, so a pull request restores the newest one, which includes `main`'s.
- The title reaches `cog` through the `PR_TITLE` environment variable and is never interpolated into the script, so a crafted title cannot run as code. `cog` runs without `--ignore-fixup-commits` or `--ignore-merge-commits`, because the squash merge makes the title the commit on `main`. The job sets a git identity first, because `cog` reads the committer from git config.
- `make drift-check` marks new files as intent-to-add before it diffs, so `git diff` sees them. Run locally, it does that to your untracked files too, so it is meant for a clean checkout like CI's.
- The runner image is pinned to `ubuntu-24.04` rather than `ubuntu-latest`, so GitHub moving `latest` to a new release cannot break CI unannounced ([upgrading.md](upgrading.md#runner-image)).
