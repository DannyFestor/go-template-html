# Postgres migration tools and data-access options

Research for issue #4 (map #1). Facts and trade-offs only; no decision is made here. Versions and activity checked 2026-10-09 through the GitHub API and each project's own repo or docs. Standing constraints: Go 1.27, stdlib-first, Postgres on Docker, Laravel-like migration commands (make:migration, migrate, rollback, status, fresh, seed), tools pinned via `go.mod` `tool` directives.

## Pinning tools: how `tool` directives work

- Since Go 1.24, `go get -tool <pkg>` adds a `tool` directive and `require`; run with `go tool <last-path-element>`; `go install tool` installs all ([go.dev: managing dependencies](https://go.dev/doc/modules/managing-dependencies)).
- Tool requirements join the module graph and MVS, so tools such as goose, sqlc and atlas add their dependency trees to `go.sum` (the docs state they "participate in minimal version selection").
- None of the READMEs/docs below document the `tool` directive; they show `go install ...@version`. All are `package main` commands and should be pinnable that way, but this was not tested here.
- golang-migrate's CLI needs build tags to include drivers (`go install -tags 'postgres' .../cmd/migrate@...`, [cmd/migrate README](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md)). Whether `go tool` can apply build tags was not checked.

## Migration tools

### Snapshot

| Tool | Latest release (date) | Last push | License | Go in go.mod | Source |
|---|---|---|---|---|---|
| goose | v3.28.0 (2026-09-02) | 2026-10-03 | MIT | 1.26.0 | [repo](https://github.com/pressly/goose) |
| golang-migrate | v4.20.1 (2026-09-09) | 2026-09-09 | MIT | 1.25.11 | [repo](https://github.com/golang-migrate/migrate) |
| Atlas | v1.3.0 (2026-08-02) | 2026-10-08 | Apache-2.0 (core) | n/a | [repo](https://github.com/ariga/atlas) |
| tern | v2.4.3 (2026-08-23) | 2026-08-23 | MIT | not checked | [repo](https://github.com/jackc/tern) |
| bun/migrate (part of bun) | v1.3.0 (2026-10-03) | 2026-10-03 | BSD-2-Clause | 1.24.0 | [repo](https://github.com/uptrace/bun) |

All five are actively maintained as of today. Release dates, stars and license come from `gh api repos/<owner>/<repo>` and `.../releases/latest`.

### goose (pressly)

Source: [README](https://github.com/pressly/goose/blob/main/README.md), [provider.go](https://github.com/pressly/goose/blob/main/provider.go), [provider_options.go](https://github.com/pressly/goose/blob/main/provider_options.go), [lock/postgres.go](https://github.com/pressly/goose/blob/main/lock/postgres.go).

- **Formats**: SQL files (`-- +goose Up` / `-- +goose Down` sections) and Go migrations as plain functions. Statements run in a transaction by default; an opt-out annotation exists for statements such as `CREATE DATABASE`. Environment variable substitution in SQL is a listed feature.
- **CLI commands**: `up`, `up-by-one`, `up-to`, `down`, `down-to`, `redo`, `reset`, `status`, `version`, `create NAME [sql|go]`, `env`, `fix`, `validate`. `create` makes a timestamped file (`-s` for sequential). Env vars `GOOSE_DRIVER`, `GOOSE_DBSTRING`, `GOOSE_MIGRATION_DIR` are supported; the CLI loads `.env` best-effort.
- **Laravel mapping**: make:migration = `create`; migrate = `up`; rollback = `down` (one step) / `down-to`; status = `status`; `reset` rolls back all. There is **no `fresh`** (drop all objects, then migrate); `reset` + `up` is the nearest, and it depends on every down migration being correct. The README lists "Seeding data" as a feature but does not describe a seed command.
- **Embedding**: legacy global API `goose.SetBaseFS(embed.FS)`; the newer `goose.NewProvider(dialect, *sql.DB, fs.FS, opts...)` takes an `fs.FS` directly. `create` and `fix` always operate on the OS filesystem, so they run from the CLI, not from embedded files.
- **Go migrations without a custom binary**: `goose.NewGoMigration(version, up, down)` plus `WithGoMigrations(...)` on the provider registers Go migrations in the app. With the plain CLI, Go migrations need your own goose binary (README).
- **Library API for boot-time use**: `Provider.Up`, `UpByOne`, `UpTo`, `Down`, `DownTo`, `Status`, `HasPending`, `GetDBVersion`, `ApplyVersion`. `Status` and `HasPending` deliberately skip locking.
- **Locking for concurrent deploys**: provider options `WithSessionLocker` (Postgres session-level advisory lock via `lock.NewPostgresSessionLocker`; retries every 5s, default 5 min lock timeout, configurable) and `WithLocker` (table-based lock `goose_lock` with heartbeat and 30s lease, `lock.NewPostgresTableLocker`). Locks are opt-in on the provider.
- **Other**: out-of-order migrations (`WithAllowOutofOrder`), `WithIsolateDDL`, slog logger (`WithSlog`). goose's own go.mod depends on `pgx/v5` v5.10.0.

### golang-migrate

Source: [README](https://github.com/golang-migrate/migrate/blob/master/README.md), [MIGRATIONS.md](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md), [postgres.go](https://github.com/golang-migrate/migrate/blob/master/database/postgres/postgres.go), [internal/cli/main.go](https://github.com/golang-migrate/migrate/blob/master/internal/cli/main.go).

- **Formats**: paired files `{version}_{title}.up.sql` / `.down.sql`; version is a sequence number or timestamp. SQL only; no Go-function migrations.
- **CLI commands** (from the CLI source): `create`, `goto`, `up [N]`, `down [N]`, `drop`, `force`, `version`. **No `status` listing**, only `version`. `drop` removes everything in the database, the nearest to `fresh` (`drop` then `up`).
- **Embedding**: `io/fs` source (`source/iofs`) accepts an `embed.FS`; go-bindata and pkger are listed as older options. `migrate.New(sourceURL, dbURL)` or `NewWithDatabaseInstance` for library use.
- **Postgres drivers**: `database/postgres` and `database/pgx/v5` (README lists PGX v4 and v5). Postgres options via URL query: `x-migrations-table`, `x-statement-timeout`, `x-multi-statement`.
- **Locking**: Postgres driver `Lock()` runs `SELECT pg_advisory_lock($1)`; the code comment says it "will wait indefinitely until the lock can be acquired"; the id derives from database name, schema and migrations table.
- **Failure model**: only the `force VERSION` command was confirmed from the sources read; the dirty-state mechanism behind it was not read.
- **README claims**: no config files or env injection by design; v3 "DO NOT USE"; supported Go 1.25 and 1.26.
- **Install caveat**: CLI needs `-tags 'postgres'` (see tool pinning above).

### Atlas (Ariga)

Source: [repo README](https://github.com/ariga/atlas), [versioned migrations intro](https://atlasgo.io/versioned/intro), [features page](https://atlasgo.io/features).

- **Model**: declarative. Define the desired schema (HCL, SQL, or an ORM such as ent/GORM) and run `atlas migrate diff` against a **dev database** (Docker) to generate timestamped SQL files plus an `atlas.sum` integrity file. `atlas migrate apply` applies; `atlas migrate status` and `atlas migrate lint` exist ([ent docs](https://entgo.io/docs/versioned-migrations)).
- **Not a blank-stub make:migration**: files are generated by diffing. A hand-authored `migrate new` was not confirmed from the pages read.
- **Edition split** (features page): Postgres tables, columns, indexes, constraints, foreign keys, enum types and comments are **Open**. Views, materialized views, functions, stored procedures, triggers, extensions, sequences, partitions, row-level security and composite/domain types are **Pro** (paid after a 30-day trial; the page lists $9 per seat per month). Linting is partly open (destructive, data-dependent, backward-incompatible); Postgres table-locking and concurrent-index analyzers and custom rules are Pro.
- **Licensing**: the core engine is Apache-2.0 on GitHub; the pages read do not separately state the license of the distributed community binary, so verify before relying on it.
- **Down migrations, embedding in Go, locking**: not covered by the pages read (the intro links to a rollback section with no content).
- **Ecosystem**: sqlc, ent and GORM integrate with it (below). ent is "developed and maintained by the Atlas team" ([ent README](https://github.com/ent/ent)).

### tern (jackc)

Source: [GitHub README](https://github.com/jackc/tern) (read through a fetch summary, not raw).

- Postgres-only. Commands: `init`, `new <name>` (numbered files), `migrate` (with `--destination`: absolute, `+N`, `-N`, `-+N`), `status`, `override-version`, `renumber start|finish`, `code install|snapshot|compile` (for functions/views), `gengen`.
- Up and down in one file, separated by `---- create above / drop below ----`. Migrations without a down section are irreversible.
- Go library `github.com/jackc/tern/v2/migrate`: `NewMigrator` with `Migration` values, SQL and Go functions; `text/template` and Sprig in migrations and config.
- Locking: advisory lock implied by changelog entries (1.10.0 "better locking", 2.4.2 skips lock when already current); no detailed doc found.
- Embedding: the page points to `gengen` for embedding in another app; whether the library reads an `embed.FS` directly was not confirmed.

### bun migrate (uptrace)

Source: [bun migrations guide](https://bun.uptrace.dev/guide/migrations.html), [migrator.go](https://github.com/uptrace/bun/blob/master/migrate/migrator.go).

- Migrations are Go functions (`MustRegister` up/down) or `.up.sql` files discovered from `embed`; `.tx.up.sql` for transactional ones. `--bun:split` separates statements.
- Rollback reverts the last **group**, not a single migration. A failed migration is still recorded as applied.
- Migrator methods include `Init`, `Migrate`, `Rollback`, `CreateGoMigration`, `CreateSQLMigrations`, `MarkApplied`, `MissingMigrations`. Locking uses a `bun_migration_locks` table.
- bun ships a `migrate` package, not a ready CLI; you write your own commands. It brings the bun dependency.

### Laravel-parity summary (migrations)

| Laravel | goose | golang-migrate | Atlas | tern |
|---|---|---|---|---|
| make:migration | `create` (timestamp or sequential, SQL or Go) | `create` | `migrate diff` (generated) | `new` |
| migrate | `up` | `up` | `migrate apply` | `migrate` |
| rollback | `down` / `down-to` | `down [N]` | not confirmed | `migrate --destination -N` |
| status | `status` | `version` only | `migrate status` | `status` |
| fresh | none; `reset` + `up` | `drop` + `up` | none confirmed | none confirmed |
| seed | none built in | none | none | none |
| Go-code migrations | yes (provider or custom binary) | no | no | yes (library) |
| Embed in binary | `fs.FS` (provider) | `iofs` | not confirmed | library values, `gengen` |
| Concurrent-deploy lock | session advisory or table lock, opt-in | advisory lock, built in | not confirmed | advisory lock (detail undocumented) |

No tool ships `fresh` or `seed`; both would be thin template-owned commands over a library API (drop schema then `Up`; a seeder that runs after migrate). This is an observation about the tools' documented commands.

### Running on boot vs via CLI

- **On boot**: library APIs exist in goose (`Provider.Up`), golang-migrate (`migrate.New(...).Up()`), tern and bun. goose and golang-migrate offer locking so replicas starting together serialise. Embedded migrations make the app a single deployable.
- **Via CLI**: all tools except bun ship a CLI. goose's CLI cannot run Go migrations unless you build a custom binary; SQL-only migrations avoid this.
- Both modes are possible with goose: CLI for `create`/`status`, provider API for boot-time `up` from the embedded FS.

## Data access options

### Snapshot

| Option | Latest release (date) | Last push | License | Go in go.mod | Source |
|---|---|---|---|---|---|
| pgx v5 | v5.11.0 (2026-09-07) | 2026-10-05 | MIT | 1.25.0 | [repo](https://github.com/jackc/pgx) |
| sqlc | v1.31.1 (2026-04-22) | 2026-10-01 | MIT | not checked | [repo](https://github.com/sqlc-dev/sqlc) |
| sqlx | v1.4.0 (2024-04-23) | 2024-08-15; last default-branch commit 2024-05-30 | MIT | not checked | [repo](https://github.com/jmoiron/sqlx) |
| bun | v1.3.0 (2026-10-03) | 2026-10-03 | BSD-2-Clause | 1.24.0 | [repo](https://github.com/uptrace/bun) |
| ent | v0.14.6 (2026-03-23) | 2026-09-30 | Apache-2.0 | 1.24 | [repo](https://github.com/ent/ent) |
| GORM | v1.31.2 (2026-06-25) | 2026-09-14 | MIT | 1.18 | [repo](https://github.com/go-gorm/gorm) |

sqlx has had no release since April 2024 and no default-branch commit since May 2024 (GitHub API). It is not archived.

### pgx directly (v5)

Source: [README](https://github.com/jackc/pgx), [CHANGELOG](https://github.com/jackc/pgx/blob/master/CHANGELOG.md), [pkg.go.dev](https://pkg.go.dev/github.com/jackc/pgx/v5).

- A low-level driver and toolkit exposing Postgres features such as `LISTEN`/`NOTIFY` and `COPY`; also has a `database/sql` adapter (`stdlib`).
- Struct scanning helpers: `CollectRows` with `RowToStructByName`, `RowToStructByNameLax`, `RowToStructByPos`; `db` struct tags.
- Transactions: `BeginFunc` / `BeginTxFunc` commit on nil and roll back on error. `Batch`/`SendBatch` for pipelining; `NamedArgs`; `pgxpool` for pooling.
- 5.11.0 adds Go 1.27 `driver.RowsColumnScanner` support in the `database/sql` adapter, so arrays and ranges scan directly through `database/sql` on Go 1.27; minimum Go remains 1.25.
- Type safety is runtime only (SQL strings and scanning); no codegen. Testing means writing your own interfaces over the pool/tx.
- Pairs with every migration tool; goose's own go.mod already uses pgx/v5.

### sqlc

Source: [config reference v1.31.1](https://docs.sqlc.dev/en/v1.31.1/reference/config.html), [DDL how-to](https://docs.sqlc.dev/en/v1.31.1/howto/ddl.html), [transactions](https://docs.sqlc.dev/en/v1.31.1/howto/transactions.html), [install](https://docs.sqlc.dev/en/v1.31.1/overview/install.html).

- Generates type-safe Go from SQL query files; you write SQL, no ORM. Runs at generate time; generated code depends on `database/sql` or pgx.
- `sql_package`: `pgx/v5`, `pgx/v4`, `database/sql` (default `database/sql`).
- Reads the schema from migration files directly. Supported formats: atlas, dbmate, golang-migrate, goose, sql-migrate, tern; "sqlc ignores down migrations when parsing SQL files".
- Transactions: generated `DBTX` interface and `New(db DBTX)`; `WithTx(tx)` returns a tx-bound `Queries`. Same pattern for pgx/v5 with context arguments. `emit_interface` outputs a `Querier` interface (useful for mocks); `emit_methods_with_db_argument` passes DBTX per call.
- Type overrides are configurable per column type.
- Install: Homebrew, snap, `go install ...@latest` (Go 1.21+), Docker, binaries; the docs do not mention `go tool`.

### sqlx

- Extension over `database/sql` (struct scanning, named queries). Maintenance stalled per the dates above. For pgx users its struct scanning overlaps pgx's `CollectRows`/`RowToStructByName`.

### bun

Source: [README](https://github.com/uptrace/bun), [docs](https://bun.uptrace.dev/guide/migrations.html).

- "SQL-first" ORM on `database/sql`: query builder with struct models and relations; PostgreSQL, MySQL, MSSQL, SQLite, Oracle. README lists migrations, fixtures, soft deletes and OpenTelemetry.
- Type safety from Go structs and a typed builder; no codegen. Its migrate package does not read goose or golang-migrate files (migrations covered above).

### ent

Source: [README](https://github.com/ent/ent), [versioned migrations](https://entgo.io/docs/versioned-migrations).

- Schema as Go code, full code generation producing a statically typed query API with graph traversal. Maintained by the Atlas team; pre-v1 (v0.14.6).
- Migrations: auto-migration or versioned migration through Atlas (`atlas migrate diff --to "ent://ent/schema"`), which can emit Atlas, golang-migrate, goose, dbmate, Flyway or Liquibase formats.
- Atlas's Pro-only Postgres objects (views, triggers, functions) matter if the schema uses them.
- Generated code and generator are a large footprint relative to the stdlib-first constraint.

### GORM

Source: [README](https://github.com/go-gorm/gorm), [migration docs](https://gorm.io/docs/migration.html).

- Full-featured reflection-based ORM: associations, hooks, preloading, transactions with nested savepoints, upsert, locking, plugins.
- `AutoMigrate` adds tables, columns, indexes, constraints; it does not delete unused columns ("WON'T delete unused columns to protect your data"). Docs say you may later need versioned migrations and point to Atlas with a GORM provider (`atlas migrate diff --env gorm`).
- Type safety is runtime (struct tags, string conditions) unless you add the separate [Gen](https://gorm.io/gen/index.html) codegen tool.
- go.mod declares Go 1.18.

### Pairing matrix

| Data layer | goose | golang-migrate | Atlas | tern |
|---|---|---|---|---|
| sqlc | schema read from goose files (up only) | read directly | read directly | read directly |
| pgx only | independent | independent | independent | independent |
| sqlx | independent | independent | independent | independent |
| bun | own migrate package; goose independent | independent | independent | independent |
| ent | Atlas can emit goose format | Atlas can emit golang-migrate format | native | not listed by ent docs |
| GORM | independent | independent | official GORM provider | independent |

"Independent" means no tool-specific integration was found in the sources read; both just talk to Postgres.

## Open questions for grilling tickets (not decided here)

- Should `fresh` and `seed` be template-owned commands over a migration library, since no tool provides them?
- SQL-only migrations (works with every tool and sqlc) versus allowing Go migrations (goose provider or tern only).
- Run migrations on boot with a lock, or only through an explicit command?
- SQL-first codegen (sqlc) versus hand-written pgx versus an ORM, against the "dependency must replace substantial code" rule.
- Do the template's needs (Postgres functions, views, triggers) stay within Atlas's open feature set, if Atlas is considered?
- Whether `go tool` can pin each CLI (goose, sqlc, atlas, golang-migrate with build tags) was not tested.
