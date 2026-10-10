# Hand-rolled Postgres job queue with leased, at-least-once delivery

The job queue is a few hundred lines of our own code on a Postgres `jobs` table, not a library. Enqueue joins the transaction carried in `ctx`, so a user row and its verification mail commit or roll back together, and jobs are as durable as the data. Workers claim due rows with `FOR UPDATE SKIP LOCKED` and hold a fixed per-kind lease (default 5 minutes, no heartbeat). A rescuer releases expired leases, so a job may run twice and every handler must tolerate that. The worker runtime (polling, goroutine pool, retries, rescue, periodic scheduling, shutdown) depends only on a consumer-declared `jobStore`. The Postgres adapter implements both that and the enqueuer, so another backend means only a new adapter.

## Considered Options

- **River**: the only library with transactional enqueue on a `pgx.Tx`, but it's open-core, with dead-letter queues and durable periodic jobs in a paid Pro tier. The template avoids open-core dependencies.
- **asynq on Valkey**: MIT, with uniqueness, a scheduler and archived jobs, but no transactional enqueue without an outbox table, durability tied to Valkey persistence (AOF is off by default) and no official Valkey support.
- **gue, neoq, goqite, asynqpg**: they take a `*sql.Tx` or `*sqlx.Tx`, always open their own transaction, or run their own migrations at startup. go-pgxq accepts a `pgx.Tx` but has a single maintainer.
- **Holding the job's transaction open while it runs** (gue's approach): crash recovery comes free, but each running job ties up a pool connection and an open transaction, which bloats MVCC during slow SMTP calls.
- **LISTEN/NOTIFY**: lower latency, but it costs a dedicated connection and pgxlisten. Polling every second is fast enough for mail.

## Consequences

- A completed job is deleted right away, because mail payloads carry live reset and verification URLs. A failed job keeps its last error for `QUEUE_FAILED_RETENTION` (default 7 days) and is reported through `ErrorReporter`.
- Uniqueness exists only for periodic jobs, through a unique `(kind, scheduled_for)` that lets every worker schedule without a leader. A project that needs unique jobs adds a partial unique index.
- A job that is cancelled at shutdown goes back to the queue without using up an attempt.
- Feature tests run due jobs synchronously on the real adapter through the harness, so transactional enqueue is exercised and not stubbed.
