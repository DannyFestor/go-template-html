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
- Make 3.81 has no `.ONESHELL` and no `--output-sync`.
- Make 3.81 reads an escaped colon in a prerequisite (`a\:b`) literally, backslash included, so it never matches the target. Target names use hyphens instead of colons.

## Pinned binaries

| | |
| --- | --- |
| Version | `tools/<tool>.version`, with `tools/<tool>.sha256` in `sha256sum` format |
| Role | Tools that aren't Go modules are downloaded into `bin/` by a real Make file target, defined once as the `pinned-binary` macro and implemented by `scripts/install-pinned-binary.sh`. |

- The target detects OS and architecture, downloads with `curl`, refuses to install an asset whose checksum doesn't match, and re-fetches when the version file, the checksum file or the script changes.
- Only macOS and Linux on arm64 and x86_64 are listed in the checksum files. Other platforms fail with "no checksum".
- Make 3.81 compares timestamps to the second, so a version file edited in the same second as the last download doesn't trigger a re-fetch.

## golangci-lint

| | |
| --- | --- |
| Version | v2.14.0, alone in `tools/golangci-lint.mod` |
| Role | Lints (`make lint`) and formats (`make fmt`, `make fmt-check`) Go code. Configured in `.golangci.yml` ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)). |

- It runs through `go tool -modfile=tools/golangci-lint.mod`, so the project toolchain compiles it and it never fails with "built with an older Go". The first run compiles it and is slow.
- `linters.default: all` turns on every linter, including deprecated ones that have a successor. Each disabled linter has its reason next to it.
- By default only one issue is shown per line. `--uniq-by-line=false` shows all of them.
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
- It reads `_test.go` files like any other file, so test-only libraries are allowed everywhere here and kept out of non-test code by depguard. `testdata/` directories are excluded.
- deepScan is off ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)). Turning it back on needs a `golang.org/x/tools` newer than the v0.49 go-arch-lint v1.19.0 requires: that version cannot read Go 1.27 export data, and deepScan then silently checks nothing.

## Architecture tests

| | |
| --- | --- |
| Version | stdlib `go/parser` and `go/ast` |
| Role | Check code shape in `tests/architecture/`: the Action shape, no exported interfaces in `internal/`, no functional options, no `go` statements in the core layers. They run with `make test` and need no Docker. |

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
| Role | Checks that commit messages are Conventional Commits (`make lint-commit-msg COMMIT_MSG_FILE=<path>`). |

- It uses the default commit types. There is no `cog.toml`.
- `fixup!` and merge commits pass, and `#` comment lines in the message file are ignored.
- Linux x86_64 gets the musl build and Linux aarch64 the gnu build, because those are the only Linux builds upstream ships for each.

## gitleaks

| | |
| --- | --- |
| Version | v8.30.1, alone in `tools/gitleaks.mod` |
| Role | Scans the staged changes for secrets (`make lint-secrets`). Configured in `.gitleaks.toml`. |

- `.gitleaks.toml` extends the default rules and allowlists test fixtures by path (`*_test.go`, `testdata/`), so a TOTP seed in a test passes.
- Findings are printed with the secret redacted, as rule, file and line.
- A false positive outside the allowlisted paths takes a `gitleaks:allow` comment on its line.

## Hooks

| Hook | Runs | Tool |
| --- | --- | --- |
| pre-commit | Refuses commits on `main`. A detached HEAD, as during a rebase, passes. | `make refuse-commit-on-main` |
| pre-commit | Formats the staged Go files and re-stages them | `make fmt` (golangci-lint) |
| pre-commit | Lints the staged shell scripts | `make lint-shell` (shellcheck) |
| pre-commit | Scans the staged changes for secrets | `make lint-secrets` (gitleaks) |
| commit-msg | Checks the message is a Conventional Commit | `make lint-commit-msg` (cog) |
| pre-push | Refuses pushes to `main`, because GitHub doesn't enforce rulesets on private repos of free accounts | `make refuse-push-to-main` |
| pre-push | Lints the Go packages changed since the merge-base with `origin/main` | `make lint` (golangci-lint) |
| pre-push | Checks the architecture when Go, `go.mod` or `.go-arch-lint.yml` changed | `make arch` (go-arch-lint) |
| pre-push | Runs the unit and architecture tests when a `.go`, `.templ`, `.sql`, `go.mod` or `go.sum` file changed | `make test` |

pre-commit and pre-push are piped: the first failing job stops the rest. No hook runs integration, feature or e2e tests. CI runs those.

## CI

| Job | Runs |
| --- | --- |
