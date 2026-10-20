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
DELETE FROM remember_tokens
WHERE
  selector = @selector
  AND validator_digest = @validator_digest
  AND expires_at > sqlc.arg(now)::timestamptz
RETURNING
  user_id;

-- name: ForgetRememberToken :exec
DELETE FROM remember_tokens
WHERE
  selector = @selector;

-- name: DeleteExpiredRememberTokens :exec
DELETE FROM remember_tokens
WHERE
  expires_at <= sqlc.arg(now)::timestamptz;
