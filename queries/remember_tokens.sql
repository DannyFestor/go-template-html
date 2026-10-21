-- name: IssueRememberToken :exec
INSERT INTO
  remember_tokens (selector, validator_digest, user_id, expires_at)
VALUES
  (
    @selector,
    @validator_digest,
    @user_id,
    @expires_at
  );

-- name: ConsumeRememberToken :one
WITH
  presented AS (
    SELECT
      candidate.selector,
      candidate.replaced_at
    FROM
      remember_tokens AS candidate
    WHERE
      candidate.selector = sqlc.arg(selector)::text
      AND candidate.validator_digest = sqlc.arg(validator_digest)::bytea
      AND candidate.expires_at > sqlc.arg(now)::timestamptz
      AND (
        candidate.replaced_at IS NULL
        OR candidate.replaced_at > sqlc.arg(replayable_after)::timestamptz
      )
    FOR UPDATE
  )
UPDATE remember_tokens
SET
  replaced_at = coalesce(
    remember_tokens.replaced_at,
    sqlc.arg(now)::timestamptz
  )
FROM
  presented
WHERE
  remember_tokens.selector = presented.selector
RETURNING
  remember_tokens.user_id,
  (presented.replaced_at IS NOT NULL)::boolean AS replay;

-- name: ForgetRememberToken :exec
DELETE FROM remember_tokens
WHERE
  selector = @selector;

-- name: ForgetRememberTokensOfUser :exec
DELETE FROM remember_tokens
WHERE
  user_id = @user_id;

-- name: DeleteExpiredRememberTokens :exec
DELETE FROM remember_tokens
WHERE
  expires_at <= sqlc.arg(now)::timestamptz
  OR replaced_at <= sqlc.arg(replayable_after)::timestamptz;
