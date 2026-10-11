# Upgrading

One section per pinned item: where it is pinned, how to bump it and what to run afterwards.

## Go

- **Pinned in**: the `go` and `toolchain` lines of `go.mod`.
- **Bump**: `go mod edit -go=<version> -toolchain=go<patch-version>`.
- **Afterwards**: `make build`.
- Go and the two linters move together in one commit ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)).
