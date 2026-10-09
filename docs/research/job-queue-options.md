# Research: job queue options for Go

Ticket: #23. Question: which Go job queue options fit this template, and what does each cost?

Researched 2026-10-09. Candidates: [River](https://github.com/riverqueue/river) (Postgres), [asynq](https://github.com/hibiken/asynq) (Redis), and a hand-rolled queue on [Redis Streams](https://redis.io/docs/latest/develop/data-types/streams/) via [go-redis](https://github.com/redis/go-redis). This document reports facts only. The choice is left to a later grilling ticket.

Source code was read at River `master@f1d9692` (2026-10-08) and asynq `master@d135f14` (2026-06-12, the latest commit). Line links point at those commits. The asynq lease, archive, shutdown and backoff constants were also checked at the v0.26.0 tag, and they match.

## Comparison table

| Criterion | River (Postgres) | asynq (Redis) | Hand-rolled on Redis Streams |
|---|---|---|---|
| Durability | Postgres WAL, so a job is as durable as any committed row | Only as durable as Redis persistence. Stock `redis.conf` is `appendonly no` plus RDB snapshots | Same as asynq: depends on Redis persistence and replication |
| Crash recovery | Rescuer reschedules jobs stuck in `running` after 1 h by default | Lease of 30 s. The recoverer retries or archives tasks whose lease expired | PEL plus `XAUTOCLAIM` (you write the reaper) |
| Retries / backoff | 25 attempts by default. Backoff is `attempt^4` s ±10 % jitter. Per-worker `NextRetry` override | 25 retries by default. Sidekiq formula `n^4 + 15 + rand(30)*(n+1)` s. `RetryDelayFunc` override | None built in. Delivery counter only |
| Failed-job storage | `discarded` rows kept 7 days by default, or forever with `-1`. Free River UI, `JobList` API | `archived` set capped at 10,000 tasks / 90 days (hard-coded). Inspector, CLI, asynqmon UI | You build a dead-letter stream (Redis docs pattern) |
| Scheduled / delayed | `InsertOpts.ScheduledAt`. Periodic jobs in OSS. "Durable periodic jobs" are Pro-only | `ProcessAt` / `ProcessIn`. Periodic tasks via `Scheduler` | None. Streams are append-only (you build a ZSET + mover) |
| Uniqueness | Partial unique index on `river_job`, enforced by the database. Duplicate insert returns the existing job | Redis lock with a TTL. The wiki calls it "best-effort" | `SET NX` (you build it). Redis 8.6+ `XADD IDMP` dedupes producer retries only |
| Graceful shutdown | `Stop`, `StopAndCancel`, `SoftStopTimeout` (since 0.38.0) | `Shutdown` with an 8 s default `ShutdownTimeout`. Unfinished tasks go back to pending. `TSTP` stops fetching | You build it. Redis 8.8 `XNACK SILENT` releases pending messages |
| Concurrency model | Goroutine per job, capped by `MaxWorkers` per queue. Fetch uses `FOR UPDATE SKIP LOCKED`. LISTEN/NOTIFY wakeups with pgx | Goroutine per task, capped by `Concurrency` (default `NumCPU`). Weighted or strict queue priority | Consumer groups. You choose the goroutine model |
| Transactional enqueue | Yes: `InsertTx(ctx, tx, …)` with `pgx.Tx` (pgx driver) or `*sql.Tx` (database/sql driver) | No. Redis cannot join a Postgres transaction | No (same reason) |
| Migrations | 8 embedded SQL migrations. CLI, `rivermigrate` Go API, or `river migrate-get` SQL dump. Documented goose recipe | None | None |
| Third-party deps (compile check, see below) | pgx, puddle, tidwall/gjson+sjson, x/sync, x/text | go-redis, uuid, robfig/cron, spf13/cast, protobuf, x/sys, x/time | go-redis only |
| Licence | MPL-2.0. River UI MPL-2.0. River Pro is proprietary | MIT. asynqmon MIT | go-redis BSD-2-Clause. Redis server ≥ 8.0 is RSALv2/SSPLv1/AGPLv3 |
| Paid tier | Pro $125/month (up to 20 devs). Enterprise custom | None (sponsorship only) | None |
| Latest release | v0.49.0, 2026-10-05 | v0.26.0, 2026-02-03 | go-redis v9.23.0, 2026-10-05 |
| Activity | 12 Go releases since 2026-07-02. 83 commits in the 30 days to 2026-10-09 (this count includes the new Java/Ruby/Rust ports) | Last commit 2026-06-12. 40 commits in the past 12 months, 0 since mid-June. Open PRs still arriving | go-redis: commits on 2026-10-08 |
| Go 1.27 | CI tests 1.27 and 1.26. `go 1.26.0` in go.mod | CI tests 1.24.x and 1.25.x only. `go 1.24.0` in go.mod | go-redis `go 1.26.0`. Compiles on 1.27.2 |
| Pre-1.0 API | v0.x. Occasional "minor breaking" changes in the changelog | v0.x. "The public API could change without a major version update" | n/a |

## River (Postgres)

### Durability and crash recovery

- Jobs are rows in `river_job` in the app's Postgres database, so they get the same durability as any committed row ([transactional enqueueing docs](https://riverqueue.com/docs/transactional-enqueueing)).
- Workers lock jobs with `FOR UPDATE SKIP LOCKED` ([river_job.sql](https://github.com/riverqueue/river/blob/f1d9692/riverdriver/riverpgxv5/internal/dbsqlc/river_job.sql#L186-L189)).
- If a worker process dies mid-job, the job stays `running` until the rescuer picks it up. `RescueStuckJobsAfter` "defaults to 1 hour, or … JobTimeout + 1 hour". The docs warn that it "can result in repeat or duplicate execution" ([client.go](https://github.com/riverqueue/river/blob/f1d9692/client.go#L379-L397)). The [graceful shutdown docs](https://riverqueue.com/docs/graceful-shutdown) say such jobs "will eventually be rescued … but not for an hour".

### Retries and backoff

- `MaxAttemptsDefault = 25` ([river_common.go](https://github.com/riverqueue/river/blob/f1d9692/internal/rivercommon/river_common.go#L16)).
- Default delay: "errorCount^4 seconds with up to 10% jitter either way" ([retrypolicy/default.go](https://github.com/riverqueue/river/blob/f1d9692/internal/retrypolicy/default.go#L42-L45)).
- You can replace the policy client-wide with `Config.RetryPolicy`, or per worker with `Worker.NextRetry` ([worker.go](https://github.com/riverqueue/river/blob/f1d9692/worker.go#L39-L57)).

### Failed-job storage and inspection

- A job that runs out of attempts moves to `discarded`. `DiscardedJobRetentionPeriod` defaults to 7 days, and "the special value -1 disables deletion" ([client.go](https://github.com/riverqueue/river/blob/f1d9692/client.go#L131-L136)).
- Completed and cancelled jobs are kept for 24 h by default ([river_shared_maintenance.go](https://github.com/riverqueue/river/blob/f1d9692/rivershared/riversharedmaintenance/river_shared_maintenance.go#L35-L37)).
- Inspection: `JobList`/`JobGet`/`JobRetry`/`JobCancel`/`JobDelete`, plus `…Tx` variants ([client.go](https://github.com/riverqueue/river/blob/f1d9692/client.go#L1529-L1667)).
- [River UI](https://github.com/riverqueue/riverui) is MPL-2.0 and was last released as v0.21.0 on 2026-10-07. The [pricing page](https://riverqueue.com/pro) lists it in the free tier.
- A "dead letter queue" is listed as a Pro feature ([pricing page](https://riverqueue.com/pro)). In OSS, failed jobs stay in `discarded`.

### Scheduled, periodic and unique jobs

- Delayed jobs: set `InsertOpts.ScheduledAt` ([insert_opts.go](https://github.com/riverqueue/river/blob/f1d9692/insert_opts.go#L58-L66)).
- `Config.PeriodicJobs` runs on the elected leader ([client.go](https://github.com/riverqueue/river/blob/f1d9692/client.go#L334-L337)). "Durable periodic jobs" are Pro-only ([pricing page](https://riverqueue.com/pro)).
- Uniqueness: "enforced with a special partial unique index on the `river_job` table". A duplicate insert sets `UniqueSkippedAsDuplicate` and returns the existing job. Uniqueness applies to insertion only, and jobs still run at least once ([unique jobs docs](https://riverqueue.com/docs/unique-jobs)).
- Uniqueness dimensions are `ByArgs` (or a subset via `river:"unique"` struct tags), `ByPeriod`, `ByQueue` and `ByState` ([insert_opts.go](https://github.com/riverqueue/river/blob/f1d9692/insert_opts.go#L102-L260)).
- Completed jobs are pruned after 24 h, after which an identical insert succeeds again ([unique jobs docs](https://riverqueue.com/docs/unique-jobs)).

### Graceful shutdown and concurrency

- `Stop` stops fetching and waits for running jobs. `StopAndCancel` cancels job contexts but still waits for the jobs to return ([client.go](https://github.com/riverqueue/river/blob/f1d9692/client.go#L1332-L1398)).
- `Config.SoftStopTimeout` turns a soft stop into a hard stop after a timeout. It was added in [0.38.0](https://github.com/riverqueue/river/blob/f1d9692/CHANGELOG.md#L226-L230) (2026-05-22).
- The docs recommend `signal.NotifyContext` → `Start(ctx)` → wait on `Stopped()`. Jobs must return an error when cancelled; otherwise "their result [would] be marked as a success" ([graceful shutdown docs](https://riverqueue.com/docs/graceful-shutdown)).
- Concurrency is set per queue with `Queues: map[string]river.QueueConfig{…: {MaxWorkers: N}}` ([doc.go](https://github.com/riverqueue/river/blob/f1d9692/doc.go#L60-L72)).
- An insert-only client is configured by omitting `Queues` and not calling `Start` ([doc.go](https://github.com/riverqueue/river/blob/f1d9692/doc.go#L74-L86)). That suits the split between the web process and `app work`.

### Transactional enqueue and the driver (pgx vs database/sql)

- `Client[TTx]` is generic over the driver's transaction type. `InsertTx(ctx, tx TTx, args, opts)` and `InsertManyTx` insert inside the caller's transaction ([client.go](https://github.com/riverqueue/river/blob/f1d9692/client.go#L1931)). `JobCompleteTx` exists for the reverse case: marking a job complete in the worker's transaction.
- `riverpgxv5` (TTx = `pgx.Tx`) is "the recommended option" ([database drivers docs](https://riverqueue.com/docs/database-drivers)).
- `riverdatabasesql` (TTx = `*sql.Tx`) works through `database/sql`, but "doesn't expose a way to implement Postgres' `LISTEN`", so clients that work jobs run in "poll only mode". The docs say it "is appropriate for clients that only insert jobs, migration tools, or installations where `LISTEN` isn't available" ([database drivers docs](https://riverqueue.com/docs/database-drivers)). Its go.mod also pulls in `lib/pq` ([go.mod](https://github.com/riverqueue/river/blob/f1d9692/riverdriver/riverdatabasesql/go.mod)).
- With sqlc on `pgx/v5`, the same `pgx.Tx` returned by `Begin(ctx)` feeds both `queries.WithTx(tx)` ([sqlc transactions how-to](https://docs.sqlc.dev/en/latest/howto/transactions.html)) and `riverClient.InsertTx(ctx, tx, …)`. River's docs have no sqlc page (`/docs/sqlc` returns 404).
- ADR-0002 has the Postgres adapter carry the transaction in `ctx`. A River enqueue adapter would therefore need to read the same `pgx.Tx` from that `ctx`, so the two adapters share a context key.

### Migrations with goose and embedded `fs.FS`

- River ships its schema as SQL files embedded in the driver (`//go:embed migration/*/*.sql`) ([riverpgxv5 driver](https://github.com/riverqueue/river/blob/f1d9692/riverdriver/riverpgxv5/river_pgx_v5_driver.go#L38-L39)). The main line currently runs to version 008. Version 008 only affects SQLite ([CHANGELOG 0.48.0](https://github.com/riverqueue/river/blob/f1d9692/CHANGELOG.md#L24-L26)), and the last Postgres-affecting migration is 007 in [0.40.0](https://github.com/riverqueue/river/blob/f1d9692/CHANGELOG.md#L161-L185).
- Applied versions are tracked in River's own `river_migration` table ([migrations docs](https://riverqueue.com/docs/migrations)).
- There are three ways to apply them ([migrations docs](https://riverqueue.com/docs/migrations)):
  1. **River CLI:** `river migrate-up` / `migrate-down` / `migrate-list`.
  2. **Go API:** `rivermigrate.New(driver, cfg).Migrate/MigrateTx` ([river_migrate.go](https://github.com/riverqueue/river/blob/f1d9692/rivermigrate/river_migrate.go#L116)). `AllVersions()` / `GetVersion(n)` return `Migration{Version, SQLUp, SQLDown}` structs ([river_migrate.go](https://github.com/riverqueue/river/blob/f1d9692/rivermigrate/river_migrate.go#L43-L53)).
  3. **SQL dump:** `river migrate-get --version N --up|--down` writes raw SQL "so it can be imported elsewhere".
- **Goose Go-migration recipe (documented):** register a goose Go migration with `goose.AddMigrationNoTxContext` that calls `rivermigrate` through `riverdatabasesql`, pinned to a specific River `TargetVersion` ([migrations docs](https://riverqueue.com/docs/migrations)). Each River upgrade that adds a migration then needs a new goose migration.
- **Alternative:** commit `migrate-get` SQL dumps as ordinary goose `.sql` files, which then live in the app's embedded FS. Goose takes migrations from an `fs.FS`, via `goose.SetBaseFS` or `goose.NewProvider(dialect, db, fsys, …)`, and can mix them with registered Go migrations (`WithGoMigrations`) ([goose.go](https://github.com/pressly/goose/blob/main/goose.go#L30-L33), [provider.go](https://github.com/pressly/goose/blob/main/provider.go#L58), [provider_options.go](https://github.com/pressly/goose/blob/main/provider_options.go#L149-L170)). Goose v3.28.0 was released 2026-09-02 and is MIT-licensed.
- The docs note that the full set of River migrations "cannot be executed within a single transaction" because of a Postgres restriction ([migrations docs](https://riverqueue.com/docs/migrations)).

### Fit behind a consumer-declared `jobEnqueuer`

- Job args must implement `Kind() string`. Workers implement `river.Worker[T]` with `Work(ctx, *river.Job[T]) error` ([job.go](https://github.com/riverqueue/river/blob/f1d9692/job.go#L21-L27), [worker.go](https://github.com/riverqueue/river/blob/f1d9692/worker.go#L39-L64)). Handlers written against app-owned types would need a thin generic wrapper in the adapter to become River workers. Otherwise `internal/job/<area>` would import River.
- Test helpers: `rivertest.RequireInserted…` asserts inserts. `rivertest.NewWorker(…).Work(ctx, tb, tx, args, opts)` runs one worker synchronously in a test transaction ([rivertest/worker.go](https://github.com/riverqueue/river/blob/f1d9692/rivertest/worker.go#L24-L104)). River has no built-in "run inline" client mode, so a synchronous dev/test adapter would be app code.
- With transactional enqueue, a job becomes visible only after commit. A synchronous adapter that runs the handler at `Enqueue` time runs it before commit. The two adapters therefore differ in ordering semantics.

### Cost, licence, maintenance, Go version

- Licence: [MPL-2.0](https://github.com/riverqueue/river/blob/f1d9692/LICENSE). River Pro is a private Go module under a proprietary EULA ([licence](https://riverqueue.com/pro/license)). Pro costs $125/month (yearly option, up to 20 developers); Enterprise is custom. Pro-only features: workflows, batching, sequences, concurrency limits, dead letter queue, durable periodic jobs, encrypted jobs, ephemeral jobs ([pricing](https://riverqueue.com/pro)).
- Releases: [v0.49.0](https://github.com/riverqueue/river/releases/tag/v0.49.0) on 2026-10-05, v0.48.0 on 2026-09-30, and 12 Go releases since v0.40.0 (2026-07-02). Last push was 2026-10-08. The repo is not archived. It has 65 open issues+PRs and about 5.8k stars.
- Recent commits are adding [Java/Ruby/Rust ports to the same repo](https://github.com/riverqueue/river/commits/master), so commit counts overstate Go-specific activity.
- The changelog still records occasional "minor breaking change" entries in v0.x, for example `rivermigrate.Validate` in [0.39.0](https://github.com/riverqueue/river/blob/f1d9692/CHANGELOG.md#L207-L217).
- Go: `go 1.26.0`, toolchain `go1.26.6` ([go.mod](https://github.com/riverqueue/river/blob/f1d9692/go.mod)). CI runs Go 1.27 and 1.26 against Postgres 14–18 ([ci.yaml](https://github.com/riverqueue/river/blob/f1d9692/.github/workflows/ci.yaml#L80-L96)). The [development doc](https://github.com/riverqueue/river/blob/f1d9692/docs/development.md) says each suite "runs on both supported Go versions".

## asynq (Redis)

### Durability and crash recovery

- Tasks live only in Redis. Redis requires 4.0+ ([README](https://github.com/hibiken/asynq#quickstart)). "Some of the lua scripts … may not be compatible with Redis Cluster" ([README](https://github.com/hibiken/asynq#redis-cluster-compatibility)).
- Durability is therefore whatever Redis persistence is configured to give. Stock `redis.conf` ships with [`appendonly no`](https://raw.githubusercontent.com/redis/redis/unstable/redis.conf) and RDB snapshots (`save 3600 1 300 100 60 10000`).
- The Redis docs say RDB users "should be prepared to lose the latest minutes of data". With AOF `appendfsync everysec` (the default once AOF is on) "you may lose 1 second of data" ([persistence docs](https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/)).
- Async replication and Sentinel failover can drop acknowledged writes ([streams docs, "Persistence, replication and message safety"](https://redis.io/docs/latest/develop/data-types/streams/#persistence-replication-and-message-safety)).
- Crash recovery: workers hold a 30 s lease (`LeaseDuration`, [rdb.go](https://github.com/hibiken/asynq/blob/d135f14/internal/rdb/rdb.go#L26)). The recoverer retries or archives tasks whose lease expired, with `ErrLeaseExpired` ([recoverer.go](https://github.com/hibiken/asynq/blob/d135f14/recoverer.go#L80-L104)).
- The same Redis would also hold the template's scs sessions. The behaviour under `maxmemory` eviction policies is not covered in asynq's docs (not verified).

### Retries, failed tasks, scheduling, uniqueness

- Retries: 25 by default, after which the task "will be moved to the **archive**". `MaxRetry`, `RetryDelayFunc`, `IsFailure` (non-failure errors don't consume a retry) and `SkipRetry` are configurable ([Task Retry wiki](https://github.com/hibiken/asynq/wiki/Task-Retry)).
- `DefaultRetryDelayFunc` is `n^4 + 15 + rand(30)*(n+1)` seconds, "taken from … sidekiq" ([server.go](https://github.com/hibiken/asynq/blob/d135f14/server.go#L401-L405)).
- Archive limits are hard-coded: `maxArchiveSize = 10000` and `archivedExpirationInDays = 90` ([rdb.go](https://github.com/hibiken/asynq/blob/d135f14/internal/rdb/rdb.go#L952-L953)).
- Completed tasks are deleted immediately unless enqueued with `Retention` ([Task Retention wiki](https://github.com/hibiken/asynq/wiki/Task-Retention-and-Result)).
- Inspection: `Inspector` API, `asynq` CLI, and the [asynqmon](https://github.com/hibiken/asynqmon) web UI (MIT). asynqmon's last release was v0.7.1 on 2022-05-06 and its last push was 2024-05-21.
- Scheduling: `ProcessAt` / `ProcessIn`. Task states are Scheduled → Pending → Active → Retry/Archived/Completed ([Life of a Task wiki](https://github.com/hibiken/asynq/wiki/Life-of-a-Task)). Periodic tasks run via `Scheduler` ([Periodic Tasks wiki](https://github.com/hibiken/asynq/wiki/Periodic-Tasks)).
- Uniqueness, two ways ([Unique Tasks wiki](https://github.com/hibiken/asynq/wiki/Unique-Tasks)):
  - `TaskID`: a conflict returns `ErrTaskIDConflict` while the task exists.
  - `Unique(ttl)`: a lock on type+payload+queue. This is "**best-effort uniqueness** … possible to enqueue a duplicate task if the lock has expired".

### Graceful shutdown and concurrency

- `Run` handles signals: `TSTP` stops fetching, `TERM`/`INT` shut down ([Signals wiki](https://github.com/hibiken/asynq/wiki/Signals)). `Shutdown()` waits `Config.ShutdownTimeout`, which defaults to 8 s ([server.go](https://github.com/hibiken/asynq/blob/d135f14/server.go#L198-L202), [L416](https://github.com/hibiken/asynq/blob/d135f14/server.go#L416)). Tasks that haven't finished by then go "back to *pending* state". `Start` exists for non-blocking use ([server.go](https://github.com/hibiken/asynq/blob/d135f14/server.go#L663-L680)).
- Concurrency: `Config.Concurrency` defaults to the CPU count ([server.go](https://github.com/hibiken/asynq/blob/d135f14/server.go#L97-L102)), with one goroutine per task. Queues can use weighted or strict priority ([README](https://github.com/hibiken/asynq#features)).

### Transactional enqueue

- Not available. `Client.Enqueue(task, opts…)` / `EnqueueContext(ctx, task, opts…)` write to Redis on their own ([client.go](https://github.com/hibiken/asynq/blob/d135f14/client.go#L369-L385)). They cannot commit atomically with a Postgres write. A transactional guarantee would need an app-built outbox table in Postgres plus a relay.

### Fit behind a consumer-declared `jobEnqueuer`

- Tasks are `NewTask(typeName string, payload []byte, opts…)`. Handlers implement `ProcessTask(ctx, *asynq.Task) error` ([README](https://github.com/hibiken/asynq#quickstart), [Handler Deep Dive wiki](https://github.com/hibiken/asynq/wiki/Handler-Deep-Dive)). The adapter owns payload (de)serialisation.
- `NewClientFromRedisClient(redis.UniversalClient)` and `NewServerFromRedisClient` accept an existing go-redis client, so the session store's client can be shared ([client.go](https://github.com/hibiken/asynq/blob/d135f14/client.go#L48), [server.go](https://github.com/hibiken/asynq/blob/d135f14/server.go#L444)).
- No in-process or synchronous mode was found. A sync adapter would be app code.

### Cost, licence, maintenance, Go version

- Licence: MIT. There is no paid tier; the README asks production users to sponsor.
- Releases: [v0.26.0](https://github.com/hibiken/asynq/releases/tag/v0.26.0) on 2026-02-03. Before that, v0.25.1 (2024-12-11) and v0.25.0 (2024-11-01).
- Activity: the last commit on `master` was 2026-06-12. There are 40 commits in the 12 months to 2026-10-09 and none since 2026-06-12. Features merged in May 2026, such as batch enqueue ([#1094](https://github.com/hibiken/asynq/pull/1094)), are unreleased. Open issues+PRs number 295, and new PRs were still being opened on 2026-10-08 ([#1184](https://github.com/hibiken/asynq/pull/1184)).
- README status: "moderate development", v0.x, and "the public API could change without a major version update before v1.0.0".
- Go: `go 1.24.0` in [go.mod](https://github.com/hibiken/asynq/blob/d135f14/go.mod). It requires `go-redis/v9 v9.14.1`, so an app on go-redis v9.23.0 would build asynq against a newer minor than it declares. The README says "the **last two** Go versions are supported", but CI tests only `1.24.x, 1.25.x` ([build.yml](https://github.com/hibiken/asynq/blob/d135f14/.github/workflows/build.yml#L10)). CI does not test Go 1.26 or 1.27.

## Hand-rolled queue on Redis Streams

The building blocks below come from the [Redis Streams docs](https://redis.io/docs/latest/develop/data-types/streams/). Everything not listed has to be written and maintained in the template.

- **Delivery:** `XADD` to enqueue, and `XREADGROUP` with a consumer group to deliver. Pending messages stay in the Pending Entries List until `XACK`. This gives at-least-once delivery.
- **Crash recovery:** "the server will leave the messages pending forever and assigned to the old consumer" unless they are claimed. `XAUTOCLAIM <min-idle-time>` (Redis 6.2+) claims idle entries.
- **Retries / backoff:** none built in. `XPENDING` exposes a per-message delivery counter. Backoff delays need a separate structure.
- **Dead letter:** the docs' pattern is that once the delivery counter "reaches a given large number … put such messages in another stream and send a notification". Redis 8.8 `XNACK FATAL` marks a message permanently failed, and `XNACK SILENT` releases a consumer's pending messages on shutdown without counting a delivery.
- **Scheduled / delayed:** not supported. The docs describe streams as "an append-only data structure". This would need a sorted set plus a mover, with Lua or `MULTI` for atomicity.
- **Uniqueness:** none for jobs. Redis 8.6+ `XADD IDMP <pid> <iid>` gives "at-most-once production" for producer retries within `IDMP-DURATION` (default 100 s, max 86,400 s) ([idempotency docs](https://redis.io/docs/latest/develop/data-types/streams/idempotency/)). That is not job-level uniqueness. go-redis v9.23.0 exposes it as `XAddArgs.ProducerID/IdempotentID` ([stream_commands.go](https://github.com/redis/go-redis/blob/v9.23.0/stream_commands.go)).
- **Retention:** `XADD MAXLEN` / `XTRIM`. Since Redis 8.2 the `ACKED` option trims only entries acknowledged by all groups. The default `KEEPREF` trims entries "regardless of whether they are referenced by any consumer groups" ([XTRIM](https://redis.io/docs/latest/commands/xtrim/)).
- **Durability:** "AOF must be used with a strong fsync policy if persistence of messages is important". Async replication means "after a failover something can be missing" ([streams docs](https://redis.io/docs/latest/develop/data-types/streams/#persistence-replication-and-message-safety)).
- **Transactional enqueue:** not possible with Postgres (outbox needed, as with asynq).
- **Interface fit:** the whole queue is app code, so any interface shape and any synchronous adapter is possible. go-redis v9.23.0 has `XAutoClaim` and `XNack` ([stream_commands.go](https://github.com/redis/go-redis/blob/v9.23.0/stream_commands.go)).
- **Dependencies:** go-redis only, which the template already uses for scs sessions. go-redis v9.23.0 was released 2026-10-05, is BSD-2-Clause, and declares `go 1.26.0` ([go.mod](https://github.com/redis/go-redis/blob/v9.23.0/go.mod)).
- **Redis server licence:** BSD-3 up to 7.2, RSALv2/SSPLv1 for 7.4–7.8, and RSALv2/SSPLv1/AGPLv3 from 8.0 ([Redis licences](https://redis.io/legal/licenses/)). XNACK needs 8.8+, `IDMP` 8.6+ and `XTRIM ACKED` 8.2+. The latest Redis release is [8.10.2](https://github.com/redis/redis/releases/tag/8.10.2) (2026-09-17). The licence applies equally to asynq, since both run on the same Redis server.

## Go 1.27 compile check and dependency weight

Go 1.27.0 was released 2026-08-19, and the current releases are 1.27.2 and 1.26.9 ([go.dev/dl](https://go.dev/dl/?mode=json), [release history](https://go.dev/doc/devel/release)).

I created a throwaway module per option that imports the library, then ran `go vet` and `go build` under **go1.27.2 darwin/arm64**. All three passed. This is a compile check only. No library test suites were run, because they need Postgres or Redis.

| Option (imports) | Modules in `go list -m all` | Third-party modules linked into the binary |
|---|---|---|
| River v0.49.0 (`river`, `riverpgxv5`, `rivermigrate`) | 29 | jackc/pgx, jackc/puddle, jackc/pgpassfile, jackc/pgservicefile, tidwall/gjson, tidwall/sjson, tidwall/match, tidwall/pretty, x/sync, x/text |
| asynq v0.26.0 | 20 | redis/go-redis, google/uuid, robfig/cron, spf13/cast, google.golang.org/protobuf, x/sys, x/time, cespare/xxhash, dgryski/go-rendezvous |
| go-redis v9.23.0 (Streams) | 11 | cespare/xxhash, go.uber.org/atomic, x/sys |

pgx is already the likely data-access driver. go-redis is already in the template for sessions.

## Not verified

- No library test suite was run on Go 1.27, only compile and vet. asynq has no CI coverage on 1.26 or 1.27.
- Who maintains asynq today, and whether a v0.27 is planned: the repo has no statement on this.
- How asynq behaves when Redis `maxmemory` eviction is active and shared with session keys: no asynq doc covers it.
- Whether River's `riverdatabasesql` works with `pgx/v5/stdlib` as the `database/sql` driver for the goose recipe. The driver doc says it is "generally still powered under the hood by Pgx" ([riverdatabasesql](https://github.com/riverqueue/river/blob/f1d9692/riverdriver/riverdatabasesql/river_database_sql_driver.go#L1-L5)), but the migrations page shows only its own example.
- The River Pro price comes from the pricing page on 2026-10-09 and may change.
