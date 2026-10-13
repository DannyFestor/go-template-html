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

## lefthook

- **Pinned in**: `tools/tools.mod` and `tools/tools.sum`, shared with the other tools that are not sensitive to version bumps ([ADR-0010](adr/0010-makefile-as-the-command-surface-with-one-binary-per-process.md)).
- **Bump**: `go get -modfile=tools/tools.mod -tool github.com/evilmartians/lefthook/v2@<version>`. A new major version also changes the module path.
- **Afterwards**: `make setup`, which reinstalls the hooks, then make a commit on a branch to see them run.

## cog

- **Pinned in**: `tools/cog.version` and `tools/cog.sha256`.
- **Bump**: write the new tag to `tools/cog.version`. Upstream publishes no checksum file, so rebuild `tools/cog.sha256` from the asset digests GitHub records:

  ```sh
  gh api repos/cocogitto/cocogitto/releases/tags/$(cat tools/cog.version) \
    --jq '.assets[] | select(.name | test("-(aarch64-apple-darwin|x86_64-apple-darwin|aarch64-unknown-linux-gnu|x86_64-unknown-linux-musl)\\.tar\\.gz$")) | "\(.digest | ltrimstr("sha256:"))  \(.name)"' \
    > tools/cog.sha256
  ```

- **Afterwards**: `make bin/cog`, then commit on a branch to see the commit-msg hook run.

## gitleaks

- **Pinned in**: `tools/gitleaks.mod` and `tools/gitleaks.sum`, alone so its dependencies stay out of the other tools' graphs.
- **Bump**: `go get -modfile=tools/gitleaks.mod -tool github.com/zricethezav/gitleaks/v8@<version>`.
- **Afterwards**: read the changelog for new default rules, then stage a change and run `make lint-secrets`.

## mvdan.cc/sh

- **Pinned in**: `tools/agenthook/go.mod` and `tools/agenthook/go.sum`, the agent hooks' own module, so the app's `go.mod` never requires it. `.golangci.yml` allows its module path in `gomodguard_v2`.
- **Bump**: `go -C tools/agenthook get mvdan.cc/sh/v3@<version>`. A new major version also changes the module path, in the imports and in `.golangci.yml`.
- **Afterwards**: read the changelog for parser and expansion changes, then run `make lint-agenthook test-agenthook`.

## Vite+

- **Pinned in**: `tools/vp.version` and `tools/vp.sha256` for `bin/vp`, and the `vite-plus` catalog entry and `vite` override in `pnpm-workspace.yaml` for the project ([ADR-0013](adr/0013-vite-plus-frontend-toolchain-with-a-build-watch-dev-loop.md)).
- **Bump**: `make vp-upgrade`. It writes the latest release tag and its `vp-checksums.txt` into `tools/`, fetches the new `bin/vp`, moves the `vite-plus` devDependency to the same release, then runs `vp migrate`, which aligns the `vite` override with it and reinstalls.
- **Afterwards**: read the release notes, because newer oxlint and oxfmt versions can flag code that passed before. Then run `make fmt-js lint-js typecheck-js test-js vite-build`.

## Node

- **Pinned in**: `.node-version`.
- **Bump**: write a newer LTS release into `.node-version` by hand, one that Vite+ supports.
- **Afterwards**: `make test-js vite-build`, which makes `bin/vp` download the new Node.

## pnpm

- **Pinned in**: the `packageManager` field of `package.json`.
- **Bump**: write the new version into `packageManager` by hand. `vp migrate` leaves it alone.
- **Afterwards**: `make setup`, which makes `bin/vp` download that pnpm and install from the unchanged lockfile.
