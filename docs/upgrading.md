# Upgrading

One section per pinned item: where it is pinned, how to bump it and what to run afterwards.

## Go

- **Pinned in**: the `go` and `toolchain` lines of `go.mod`.
- **Bump**: `go mod edit -go=<version> -toolchain=go<patch-version>`.
- **Afterwards**: `make build`.
- Go and the two linters move together in one commit ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)).

## golangci-lint

- **Pinned in**: `tools/golangci-lint.mod` and `tools/golangci-lint.sum`, alone so no other tool moves its dependencies ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)).
- **Bump**: `go get -modfile=tools/golangci-lint.mod -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<version>`.
- **Afterwards**: read the changelog for new linters, which `default: all` switches on, and for deprecations, which need a disable with a reason. Then run `make fmt-check lint`.

## go-arch-lint

- **Pinned in**: `tools/go-arch-lint.mod` and `tools/go-arch-lint.sum`, alone so no other tool moves its dependencies ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)).
- **Bump**: `go get -modfile=tools/go-arch-lint.mod -tool github.com/fe3dback/go-arch-lint@<version>`.
- **Afterwards**: `make arch`.

## shellcheck

- **Pinned in**: `tools/shellcheck.version` and `tools/shellcheck.sha256`.
- **Bump**: write the new tag to `tools/shellcheck.version`. Upstream publishes no checksum file, so rebuild `tools/shellcheck.sha256` from the asset digests GitHub records:

  ```sh
  gh api repos/koalaman/shellcheck/releases/tags/$(cat tools/shellcheck.version) \
    --jq '.assets[] | select(.name | test("\\.(darwin|linux)\\.(aarch64|x86_64)\\.tar\\.gz$")) | "\(.digest | ltrimstr("sha256:"))  \(.name)"' \
    > tools/shellcheck.sha256
  ```

- **Afterwards**: `make lint-shell`, which re-fetches `bin/shellcheck`.
