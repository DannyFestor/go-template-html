# Architecture enforced by go-arch-lint, depguard stdlib denies and AST architecture tests

The rules from [ADR-0002](0002-layered-architecture-with-actions.md) are checked by three mechanisms, each covering what the others cannot. go-arch-lint owns the dependency graph between internal layers and confines each third-party library to the component that owns it; it is a whitelist, so a package outside every declared component fails the check, which keeps agents from inventing new layers. go-arch-lint always allows the standard library, so depguard holds only stdlib denies (`net/http` and `database/sql` in the core layers, `reflect` everywhere). Rules about code shape (the Action signature, no exported interfaces, no functional options, no `go` statements in core layers) have no stock linter and are stdlib `go/parser` tests in `tests/architecture/`. Golangci-lint and go-arch-lint each sit in their own `-modfile` and are compiled by the project toolchain, so a Go bump never trips golangci-lint's "built with an older Go" check; a Go bump and both tool bumps land in one commit.

## Considered Options

- **depguard alone**: per-file-glob prefix lists cannot flag a package that matches no rule, repeat the layer-to-path mapping in every rule, and match plain string prefixes (`internal/domain` also allows `internal/domainx`).
- **All import rules in the architecture tests**: one layer table in Go with whitelist semantics and no extra tool, but no dependency graph, and a second home for rules that a dedicated tool already expresses declaratively.
- **One shared tools modfile**: tools share one MVS graph, so one tool's `golang.org/x/tools` requirement moves another's; golangci-lint upstream warns against touching its dependencies.
- **Prebuilt golangci-lint binaries**: break whenever the project's Go version passes the version the binary was built with.

## Consequences

- A new layer or library needs an edit to `.go-arch-lint.yml`; a new stdlib restriction needs a depguard rule; a new shape rule or exception needs an edit in `tests/architecture/`.
- deepScan is off for `bootstrap`, because passing concrete adapters into constructors is that package's job.
- A Go release waits until golangci-lint and go-arch-lint support it.
