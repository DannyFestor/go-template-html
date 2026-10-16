-- name: InsertJob :exec
INSERT INTO
  jobs (id, kind, args, run_at, created_at)
VALUES
  (@id, @kind, @args, @run_at, @run_at);

-- name: InsertPeriodicJob :exec
INSERT INTO
  jobs (id, kind, args, run_at, scheduled_for, created_at)
VALUES
  (
    @id,
    @kind,
    @args,
    sqlc.arg(scheduled_for)::timestamptz,
    sqlc.arg(scheduled_for)::timestamptz,
    @created_at
  )
ON CONFLICT (kind, scheduled_for) DO NOTHING;

-- name: ClaimJob :one
WITH
  leases AS (
    SELECT
      unnest(@kinds::text[]) AS kind,
      unnest(@lease_ms::bigint[]) AS lease_ms
  ),
  next AS (
    SELECT
      jobs.id,
      leases.lease_ms
    FROM
      jobs
      JOIN leases ON leases.kind = jobs.kind
    WHERE
      jobs.state = 'available'
      AND jobs.run_at <= sqlc.arg(now)::timestamptz
    ORDER BY
      jobs.run_at,
      jobs.id
    LIMIT
      1
    FOR UPDATE OF
      jobs SKIP LOCKED
  )
UPDATE jobs
SET
  state = 'running',
  attempt = jobs.attempt + 1,
  locked_until = sqlc.arg(now)::timestamptz + make_interval(secs => next.lease_ms / 1000.0)
FROM
  next
WHERE
  jobs.id = next.id
RETURNING
  jobs.id,
  jobs.kind,
  jobs.args,
  jobs.attempt;

-- name: CompleteJob :exec
DELETE FROM jobs
WHERE
  id = @id
  AND attempt = @attempt
  AND state = 'running';

-- name: RetryJob :exec
UPDATE jobs
SET
  state = 'available',
  run_at = @run_at,
  locked_until = NULL,
  last_error = sqlc.arg(last_error)::text
WHERE
  id = @id
  AND attempt = @attempt
  AND state = 'running';

-- name: FailJob :exec
UPDATE jobs
SET
  state = 'failed',
  failed_at = sqlc.arg(failed_at)::timestamptz,
  locked_until = NULL,
  last_error = sqlc.arg(last_error)::text
WHERE
  id = @id
  AND attempt = @attempt
  AND state = 'running';

-- name: ReleaseJob :exec
UPDATE jobs
SET
  state = 'available',
  attempt = attempt - 1,
  locked_until = NULL
WHERE
  id = @id
  AND attempt = @attempt
  AND state = 'running';

-- name: RescueExpiredJobs :execrows
UPDATE jobs
SET
  state = 'available',
  locked_until = NULL,
  last_error = sqlc.arg(last_error)::text
WHERE
  state = 'running'
  AND locked_until <= sqlc.arg(now)::timestamptz;

-- name: DeleteFailedJobsBefore :exec
DELETE FROM jobs
WHERE
  state = 'failed'
  AND failed_at < sqlc.arg(failed_before)::timestamptz;
