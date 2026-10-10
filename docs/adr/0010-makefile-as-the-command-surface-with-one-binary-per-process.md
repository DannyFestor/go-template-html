# Makefile as the command surface, with one binary per process

Every developer command is a target in one self-documenting Makefile, and Go code exists only for what a shell recipe cannot do. Each process gets its own small `main`: `cmd/web`, `cmd/worker`, `cmd/migrate` (`up` and `status`) and `cmd/seed`. Shell recipes are less code to maintain than Go subcommands for creating migrations, resetting the dev schema or downloading pinned binaries. Make is already installed wherever Go, Docker and git are, so a fresh clone needs nothing else. Targets that only exist in Make and Compose (such as `migrate-fresh`) cannot run in production, so the Go code needs no production guard. Composite targets only call other targets, and each leaf target runs a single command. pre-push, CI and the Dockerfile call those same targets, so each command is defined once.

## Considered Options

- **One `cmd/app` binary with artisan-style subcommands** (`serve`, `work`, `migrate`, `seed`): a flag-parsing dispatcher plus Go code for jobs a shell handles, and dev-only commands such as `fresh` would ship in the production binary behind a runtime guard.
- **Taskfile via a `tool` directive**: needs no install, but adds 125 indirect requires, including the AWS and GCP SDKs.
- **mise**: pins every tool and runs tasks, but has to be installed before the repo works.
- **`go generate` as the single generation entrypoint**: templ and sqlc run once for the whole repo, not per package, and calling tools that live in a separate modfile would need hard-coded relative paths.

## Consequences

- The Makefile must run on GNU Make 3.81, the version macOS ships: no `.ONESHELL`, no `--output-sync`. `make dev` therefore interleaves the output of its parallel watchers.
- Windows is supported only through WSL2.
- `make generate` runs templ, sqlc and then `go generate ./...`. Only mockgen and `migrations.sum` use `//go:generate`.
- air, lefthook and sqlc share `tools/tools.mod`, because they are not as sensitive to version bumps as the linters ([ADR-0007](0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)). templ and mockgen stay in the main `go.mod`, which keeps templ's CLI and runtime on the same version and lets `//go:generate` call mockgen without a modfile path.
- The production image contains only `web`, `worker` and `migrate`.
