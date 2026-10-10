# Test kinds selected by build tag and path, with a database per test

Go tests come in four kinds, and each kind's Make target selects it by build tag plus path. Unit and architecture tests are untagged, so plain `go test ./...` passes without Docker. Integration tests are tagged `integration` and sit next to each adapter. Feature tests are tagged `feature` under `tests/feature/`, and e2e tests are tagged `e2e` under `tests/e2e/`. Every integration and feature test gets its own Postgres database, cloned with `CREATE DATABASE … TEMPLATE` from a template database named after the `migrations.sum` hash. A test-only harness builds that template once, under an advisory lock, and keeps it between runs. Every Valkey key is prefixed per test. This lets every test call `t.Parallel()`. Rolling back a transaction per test can't isolate feature tests, because the server handles each request on its own goroutine and connection.

## Considered Options

- **Rollback per test** (Laravel's `RefreshDatabase`): it only works when the test and the code share one connection. That excludes feature tests, and nested `WithinTransaction` calls would join the test's transaction and hide commit behaviour.
- **One database per worker** (Pest and ParaTest's `<db>_test_N`): Go runs parallel tests as goroutines inside one process, so there's no worker to own a database.
- **A schema per package via `DB_SCHEMA`**: tests within a package would still share data and need unique values and cleanup discipline.
- **testcontainers**: one more dependency and a container per run, when the dev Compose Postgres can already create databases.
- **`pgtestdb`**: the same design, but it pulls in lib/pq, go-cmp and testy. The template already depends on pgx, and `migrations.sum` already provides the hash.

## Consequences

- A failed test's database is kept and its name is logged for inspection. `make test-db-clean` removes leftover clones and stale templates.
- The Valkey adapter needs a `VALKEY_KEY_PREFIX` setting that applies to every key and pub/sub channel. That also lets several apps share one Valkey.
- `paralleltest` is switched back on. A test that can't run in parallel needs `//nolint:paralleltest // <why>`.
- Passing a tag re-runs the untagged tests in the selected paths as well. Those are fast, so this is accepted rather than worked around.
- A fifth kind, **JS unit**, runs Vitest with happy-dom on `*.test.ts` files that sit next to the TypeScript components ([ADR-0013](0013-vite-plus-frontend-toolchain-with-a-build-watch-dev-loop.md)). `make test` runs both Go unit and JS unit tests, and neither needs Docker. JS coverage is reported separately and does not count toward the merged Go coverage.
