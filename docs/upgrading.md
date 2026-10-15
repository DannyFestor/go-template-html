# Upgrading

One section per pinned item: where it is pinned, how to bump it and what to run afterwards.

## Go

- **Pinned in**: the `go` and `toolchain` lines of `go.mod`.
- **Bump**: `go mod edit -go=<version> -toolchain=go<patch-version>`, then move the [Go build image](#go-build-image) to the same patch version.
- **Afterwards**: `make build image`.
- Go and the two linters move together in one commit ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)).

## templ

- **Pinned in**: the `tool github.com/a-h/templ/cmd/templ` directive and the `github.com/a-h/templ` require in `go.mod`, which pin the CLI and the runtime together.
- **Bump**: `go get -tool github.com/a-h/templ/cmd/templ@<version>`.
- **Afterwards**: `make generate`, which rewrites the version line of every `*_templ.go`, then `make fmt-check-templ test-go vite-build test-feature`. Commit the regenerated files with the bump.

## mockgen

- **Pinned in**: the `tool go.uber.org/mock/mockgen` directive and the `go.uber.org/mock` require in `go.mod`, which pin the CLI and the `gomock` runtime together.
- **Bump**: `go get -tool go.uber.org/mock/mockgen@<version>`. A new major version also changes the module path, in the imports, `.go-arch-lint.yml` and `.golangci.yml`.
- **Afterwards**: `make generate`, which rewrites every `mock_*_test.go`, then `make test-go`. Commit the regenerated files with the bump.

## Go build image

- **Pinned in**: the first `FROM` line of `Dockerfile`, as `golang:<patch-version>-<debian-release>`. The patch version is the one on the `toolchain` line of `go.mod`.
- **Bump**: change it together with `go.mod` (see [Go](#go)). Move to a newer Debian release once Docker Hub publishes a `golang` tag for it.
- **Afterwards**: `make image`.

## Distroless base image

- **Pinned in**: the second `FROM` line of `Dockerfile`, as `gcr.io/distroless/static-debian12:nonroot@sha256:<digest>`. The digest is what pins it, so the tag stays as a label for humans.
- **Bump**: look up the digest the tag points at now and replace it:

  ```sh
  docker buildx imagetools inspect gcr.io/distroless/static-debian12:nonroot --format '{{json .Manifest.Digest}}'
  ```

- **Afterwards**: `make image`, then `docker run --rm go-template-html:latest` to check the new base still runs `/app/web`.

## Compose service images

- **Pinned in**: the `image:` lines of `compose.yaml`. Postgres, Valkey and Mailpit are pinned to a major or minor tag, which keeps receiving patch releases. pgweb and Valkey Admin publish no such tags, so they are pinned to an exact version.
- **Bump**: list the image's tags and change the tag in `compose.yaml`. Official images such as Postgres live under `library/`:

  ```sh
  curl -s 'https://hub.docker.com/v2/repositories/valkey/valkey/tags?page_size=50&ordering=last_updated' | jq -r '.results[].name'
  ```

  Check the version the image reports before you pick a tag, because a fresh minor tag can still point at a release candidate. A new Postgres major version can't read the old one's data, so run `make reset` first, which deletes it.
- **Afterwards**: `make services`, which pulls the new tags, recreates their containers and waits until all are healthy. `docker compose pull` fetches the latest patch release of tags you didn't change.

## golangci-lint

- **Pinned in**: `tools/golangci-lint.mod` and `tools/golangci-lint.sum`, alone so no other tool moves its dependencies ([ADR-0007](adr/0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)).
- **Bump**: `go get -modfile=tools/golangci-lint.mod -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<version>`.
- **Afterwards**: read the changelog for new linters, which `default: all` switches on, and for deprecations, which need a disable with a reason. Then run `make fmt-check-go lint`.

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

## air

- **Pinned in**: `tools/tools.mod` and `tools/tools.sum`, shared with lefthook ([ADR-0010](adr/0010-makefile-as-the-command-surface-with-one-binary-per-process.md)).
- **Bump**: `go get -modfile=tools/tools.mod -tool github.com/air-verse/air@<version>`.
- **Afterwards**: read the changelog for renamed flags, because `watch-assets` configures air through flags alone. Then run `make dev`, edit a `.ts` file and check that the page reloads.

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

## govulncheck

- **Pinned in**: `tools/govulncheck.mod` and `tools/govulncheck.sum`, alone so its `golang.org/x/tools` stays out of the other tools' graphs.
- **Bump**: `go get -modfile=tools/govulncheck.mod -tool golang.org/x/vuln/cmd/govulncheck@<version>`.
- **Afterwards**: `make vuln vuln-agenthook`.

## pgx

- **Pinned in**: `github.com/jackc/pgx/v5` in `go.mod` and `go.sum`.
- **Bump**: `go get github.com/jackc/pgx/v5@<version>`. A new major version also changes the module path, in the imports, `.go-arch-lint.yml` and `.golangci.yml`.
- **Afterwards**: read the changelog for changed pool defaults and connection parameters, then run `make services test-integration vite-build test-feature`.

## goose

- **Pinned in**: `github.com/pressly/goose/v3` in `go.mod` and `go.sum`. Only `internal/adapter/goose` imports it ([ADR-0005](adr/0005-forward-only-goose-migrations-with-sum-file.md)).
- **Bump**: `go get github.com/pressly/goose/v3@<version>`. A new major version also changes the module path, in the imports, `.go-arch-lint.yml` and `.golangci.yml`.
- **Afterwards**: read the changelog for changes to the `Provider` API, the version table, the session lock and how SQL files are parsed, such as new `-- +goose` annotations. Then run `make services test-integration migrate-fresh migrate-status`.

## sql-formatter

- **Pinned in**: `sql-formatter` in the `devDependencies` of `package.json`, and `pnpm-lock.yaml`.
- **Bump**: `VP_HOME="$PWD/.vite-plus" bin/vp add -D sql-formatter@<version>`.
- **Afterwards**: run `make fmt-sql` and review the diff in `queries/`. Applied migrations are never swept, so they stay byte-identical; check one by hand with `make fmt-check-sql FILES=migrations/<file>.sql` to see how the new version would format the next one.

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

## htmx

- **Pinned in**: `htmx.org` in `package.json`, at an exact version, and `pnpm-lock.yaml`.
- **Bump**: `VP_HOME="$PWD/.vite-plus" bin/vp add htmx.org@<version>` with the exact version. While htmx 4 sits under npm's `next` dist-tag, `@latest` would install 2.x.
- **Afterwards**: read the changelog for renamed events and attributes, such as `htmx:after:history:push` in `internal/asset/src/components/nav.ts` and `hx-boost:inherited` in `internal/view/layout/document.templ`. Then run `make typecheck-js test-js vite-build test-feature` and click through a boosted navigation in the browser.

## Alpine

- **Pinned in**: `@alpinejs/csp`, `@alpinejs/persist`, `@types/alpinejs` and `@types/alpinejs__persist` in `package.json`, and `pnpm-lock.yaml`.
- **Bump**: `VP_HOME="$PWD/.vite-plus" bin/vp add @alpinejs/csp@<version> @alpinejs/persist@<version>`, and the `-D` type packages to the matching minor version.
- **Afterwards**: read the release notes for the CSP build's expression support, then run `make typecheck-js test-js vite-build` and open the app layout in the browser.

## GitHub Actions

- **Pinned in**: every `uses:` line in `.github/workflows/`, by full commit SHA with the version in a trailing comment. Nothing bumps them automatically.
- **Bump**: look up the latest release and the commit its tag points at, then replace the SHA and the comment on every line that uses the action:

  ```sh
  gh api repos/actions/checkout/releases/latest --jq .tag_name
  gh api repos/actions/checkout/commits/<tag> --jq .sha
  ```

- **Afterwards**: read the release notes for changed inputs or defaults. A new major version of `actions/setup-go` or `actions/cache` can move its cache keys, so the first run after the bump is cold. Then open a pull request and watch every check.

## Runner image

- **Pinned in**: `runs-on: ubuntu-24.04` on every job in `.github/workflows/`.
- **Bump**: replace the label on every job once GitHub lists the next Ubuntu LTS image as generally available.
- **Afterwards**: open a pull request and watch every check. The image ships the system tools the Makefile relies on, such as `make`, `curl` and `git`, so a change in their versions shows up there.
