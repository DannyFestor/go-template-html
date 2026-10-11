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

## Hooks

| Hook | Runs | Tool |
| --- | --- | --- |

## CI

| Job | Runs |
| --- | --- |
