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

## shellcheck

| | |
| --- | --- |
| Version | `tools/shellcheck.version`, a [pinned binary](#pinned-binaries) in `bin/shellcheck` |
| Role | Lints every tracked `*.sh` file (`make lint-shell`). |

- Untracked scripts and scripts without the `.sh` extension are not linted.

## Hooks

| Hook | Runs | Tool |
| --- | --- | --- |

## CI

| Job | Runs |
| --- | --- |
