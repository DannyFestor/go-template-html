# sqlc-generated data access with opaque string IDs

Fixed-shape queries are plain SQL in a root `queries/` directory, next to `migrations/`, and sqlc generates typed pgx/v5 code from both into `internal/adapter/postgres/internal/sqlcgen`, where Go's `internal` rule keeps row types out of every other package; per-aggregate stores in `adapter/postgres` map those rows to domain types and driver errors to domain errors. Domain IDs are opaque per-entity string types produced by an injected `idGenerator` (UUIDv7 by default), not `uuid.UUID`, so switching to ULIDs or nanoids touches a generator adapter, the store's conversion and a migration, never `domain` or the actions.

## Considered Options

- **ORM (GORM, ent, bun)**: hides the SQL behind its own model and pulls the library past the adapter seam; ent is pre-v1 and depends on Atlas.
- **Hand-written pgx everywhere**: no codegen, but every scan and parameter list is written and kept in sync by hand.
- **Optional filters in sqlc as `(sqlc.narg('x') IS NULL OR col = sqlc.narg('x'))`**: once Postgres switches a cached prepared statement to a generic plan, the condition can no longer be folded and the index may be skipped. Queries with optional filters or user-chosen sort are instead hand-written in their store, using a store-private builder that appends only the conditions present and takes sort columns from a closed set; no query-builder dependency.
- **`uuid.UUID` as the domain ID type**: type-safe, but changing the ID scheme would reach into `domain` and every action.
- **IDs and timestamps from database defaults only**: the app supplies both so actions and tests control them through the injected generator and clock; `DEFAULT uuidv7()` and `DEFAULT now()` remain as a safety net, which requires Postgres 18 or later.

## Consequences

- Another database is a sibling adapter with its own sqlc config, migrations, queries and error mapping; nothing above the adapter changes.
- Dynamic-filter queries are the one place SQL lives in Go rather than `queries/`.
- `sqlc.yaml` overrides map `uuid` and `timestamptz` to `uuid.UUID` and `time.Time`, nullable columns to pointers; nullable overrides must use the object form because sqlc doubles the pointer in the string form for slash-less import paths.
